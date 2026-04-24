package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltHex := hex.EncodeToString(salt)
	h := sha256.New()
	h.Write([]byte(saltHex + password))
	return saltHex + ":" + hex.EncodeToString(h.Sum(nil)), nil
}

func checkPassword(password, stored string) bool {
	parts := splitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	h := sha256.New()
	h.Write([]byte(parts[0] + password))
	actual := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(actual), []byte(parts[1]))
}

func splitN(s, sep string, n int) []string {
	var parts []string
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		parts = append(parts, s[:idx])
		s = s[idx+len(sep):]
	}
	parts = append(parts, s)
	return parts
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: reset_password <username_or_email> <new_password>\n")
		os.Exit(1)
	}
	identifier := os.Args[1]
	newPassword := os.Args[2]

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getenv("DB_USER", "finances_user"),
		getenv("DB_PASSWORD", "finances_pass"),
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_NAME", "my_finances"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	hash, err := hashPassword(newPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash: %v\n", err)
		os.Exit(1)
	}

	res, err := db.Exec(
		`UPDATE users SET password_hash=$1 WHERE username=$2 OR email=$2`,
		hash, identifier,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		os.Exit(1)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fmt.Fprintf(os.Stderr, "no user found with username or email %q\n", identifier)
		os.Exit(1)
	}
	fmt.Printf("Password updated for %q\n", identifier)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
