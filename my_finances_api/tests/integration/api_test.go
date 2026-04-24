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
