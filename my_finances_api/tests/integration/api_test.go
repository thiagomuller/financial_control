package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/thiago/my_finances_api/internal/config"
	"github.com/thiago/my_finances_api/internal/database"
	"github.com/thiago/my_finances_api/internal/handlers"
)

func setupDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("starting postgres container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	cfg := &config.Config{
		DBHost:     host,
		DBPort:     port.Port(),
		DBName:     "testdb",
		DBUser:     "testuser",
		DBPassword: "testpass",
		JWTSecret:  "test-secret",
		Port:       "0",
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	return db, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	cfg := &config.Config{JWTSecret: "test-secret", Port: "8080"}
	srv := httptest.NewServer(handlers.NewRouter(db, cfg))
	defer srv.Close()

	// Register
	body, _ := json.Marshal(map[string]string{
		"name": "Test User", "username": "testuser", "email": "test@example.com", "password": "pass123",
	})
	resp, err := http.Post(srv.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", resp.StatusCode)
	}

	var authResp map[string]any
	json.NewDecoder(resp.Body).Decode(&authResp)
	resp.Body.Close()

	token, ok := authResp["token"].(string)
	if !ok || token == "" {
		t.Fatal("register: expected non-empty token")
	}

	// Login
	body, _ = json.Marshal(map[string]string{"username": "testuser", "password": "pass123"})
	resp, err = http.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestBankAccountCRUD(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	cfg := &config.Config{JWTSecret: "test-secret", Port: "8080"}
	srv := httptest.NewServer(handlers.NewRouter(db, cfg))
	defer srv.Close()

	token := registerAndGetToken(t, srv.URL, "bankuser", "bank@test.com")

	// Create
	body, _ := json.Marshal(map[string]any{"name": "Checking", "initial_balance": 1000.0})
	req, _ := http.NewRequest("POST", srv.URL+"/api/bank-accounts", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create bank account: expected 201, got %d", resp.StatusCode)
	}

	var account map[string]any
	json.NewDecoder(resp.Body).Decode(&account)
	resp.Body.Close()

	id := account["id"].(string)

	// List
	req, _ = http.NewRequest("GET", srv.URL+"/api/bank-accounts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list bank accounts: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Get
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/bank-accounts/%s", srv.URL, id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get bank account: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Delete
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/api/bank-accounts/%s", srv.URL, id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete bank account: expected 204, got %d", resp.StatusCode)
	}
}

func TestTransferInsufficientBalance(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	cfg := &config.Config{JWTSecret: "test-secret", Port: "8080"}
	srv := httptest.NewServer(handlers.NewRouter(db, cfg))
	defer srv.Close()

	token := registerAndGetToken(t, srv.URL, "transferuser", "transfer@test.com")

	srcID := createBankAccount(t, srv.URL, token, "Source", 100.0)
	tgtID := createBankAccount(t, srv.URL, token, "Target", 0.0)

	// Attempt transfer exceeding balance
	body, _ := json.Marshal(map[string]any{
		"name": "big transfer", "value": 500.0,
		"source_account_id": srcID, "target_account_id": tgtID,
	})
	req, _ := http.NewRequest("POST", srv.URL+"/api/transfers", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for insufficient balance, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestBankAccountWithImage(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	cfg := &config.Config{JWTSecret: "test-secret", Port: "8080"}
	srv := httptest.NewServer(handlers.NewRouter(db, cfg))
	defer srv.Close()

	token := registerAndGetToken(t, srv.URL, "imguser", "img@test.com")

	smallIcon := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg=="

	// Create with icon
	body, _ := json.Marshal(map[string]any{
		"name":            "Icon Account",
		"initial_balance": 0.0,
		"icon_url":        smallIcon,
	})
	req, _ := http.NewRequest("POST", srv.URL+"/api/bank-accounts", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create with icon: expected 201, got %d", resp.StatusCode)
	}
	var account map[string]any
	json.NewDecoder(resp.Body).Decode(&account)
	resp.Body.Close()
	id := account["id"].(string)

	// Update with new icon
	newIcon := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJ"
	body, _ = json.Marshal(map[string]any{"name": "Icon Account", "icon_url": newIcon})
	req, _ = http.NewRequest("PUT", fmt.Sprintf("%s/api/bank-accounts/%s", srv.URL, id), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update with new icon: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Remove icon by sending null
	body, _ = json.Marshal(map[string]any{"name": "Icon Account", "icon_url": nil})
	req, _ = http.NewRequest("PUT", fmt.Sprintf("%s/api/bank-accounts/%s", srv.URL, id), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("remove icon: expected 200, got %d", resp.StatusCode)
	}
	var updated map[string]any
	json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()
	if updated["icon_url"] != nil {
		t.Errorf("expected icon_url to be null after removal, got %v", updated["icon_url"])
	}
}

func TestBankAccountNegativeInitialBalance(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	cfg := &config.Config{JWTSecret: "test-secret", Port: "8080"}
	srv := httptest.NewServer(handlers.NewRouter(db, cfg))
	defer srv.Close()

	token := registerAndGetToken(t, srv.URL, "neguser", "neg@test.com")

	body, _ := json.Marshal(map[string]any{"name": "Bad Account", "initial_balance": -100.0})
	req, _ := http.NewRequest("POST", srv.URL+"/api/bank-accounts", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("negative initial_balance: expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestBankAccountStatementIncludesTransfers(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	cfg := &config.Config{JWTSecret: "test-secret", Port: "8080"}
	srv := httptest.NewServer(handlers.NewRouter(db, cfg))
	defer srv.Close()

	token := registerAndGetToken(t, srv.URL, "stmtuser", "stmt@test.com")

	srcID := createBankAccount(t, srv.URL, token, "Source Account", 500.0)
	tgtID := createBankAccount(t, srv.URL, token, "Target Account", 0.0)

	body, _ := json.Marshal(map[string]any{
		"name": "monthly savings", "value": 100.0,
		"source_account_id": srcID, "target_account_id": tgtID,
	})
	req, _ := http.NewRequest("POST", srv.URL+"/api/transfers", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create transfer: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Transfer appears in source statement as subtract
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/bank-accounts/%s/statement", srv.URL, srcID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get source statement: expected 200, got %d", resp.StatusCode)
	}

	var stmtResp map[string]any
	json.NewDecoder(resp.Body).Decode(&stmtResp)
	resp.Body.Close()

	txnsMap, _ := stmtResp["transactions"].(map[string]any)
	data, _ := txnsMap["data"].([]any)
	if len(data) == 0 {
		t.Fatal("expected at least one entry in source statement")
	}
	entry, _ := data[0].(map[string]any)
	if entry["kind"] != "transfer" {
		t.Errorf("source: expected kind=transfer, got %v", entry["kind"])
	}
	if entry["operation"] != "subtract" {
		t.Errorf("source: expected operation=subtract, got %v", entry["operation"])
	}

	// Transfer appears in target statement as add
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/bank-accounts/%s/statement", srv.URL, tgtID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get target statement: expected 200, got %d", resp.StatusCode)
	}

	json.NewDecoder(resp.Body).Decode(&stmtResp)
	resp.Body.Close()

	txnsMap, _ = stmtResp["transactions"].(map[string]any)
	data, _ = txnsMap["data"].([]any)
	if len(data) == 0 {
		t.Fatal("expected at least one entry in target statement")
	}
	entry, _ = data[0].(map[string]any)
	if entry["kind"] != "transfer" {
		t.Errorf("target: expected kind=transfer, got %v", entry["kind"])
	}
	if entry["operation"] != "add" {
		t.Errorf("target: expected operation=add, got %v", entry["operation"])
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func registerAndGetToken(t *testing.T, baseURL, username, email string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"name": "User", "username": username, "email": email, "password": "pass123",
	})
	resp, err := http.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var auth map[string]any
	json.NewDecoder(resp.Body).Decode(&auth)
	return auth["token"].(string)
}

func createBankAccount(t *testing.T, baseURL, token, name string, balance float64) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"name": name, "initial_balance": balance})
	req, _ := http.NewRequest("POST", baseURL+"/api/bank-accounts", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var acc map[string]any
	json.NewDecoder(resp.Body).Decode(&acc)
	return acc["id"].(string)
}
