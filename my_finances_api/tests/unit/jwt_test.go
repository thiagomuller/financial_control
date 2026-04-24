package unit_test

import (
	"testing"

	"github.com/thiago/my_finances_api/internal/auth"
)

func TestGenerateAndValidateToken(t *testing.T) {
	cases := []struct {
		name    string
		userID  string
		secret  string
		wantErr bool
	}{
		{"valid token", "user-123", "my-secret", false},
		{"different user id", "user-456", "my-secret", false},
		{"short secret", "u1", "x", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tc.userID, tc.secret)
			if (err != nil) != tc.wantErr {
				t.Fatalf("GenerateToken() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			got, err := auth.ValidateToken(token, tc.secret)
			if err != nil {
				t.Fatalf("ValidateToken() unexpected error: %v", err)
			}
			if got != tc.userID {
				t.Errorf("ValidateToken() = %q, want %q", got, tc.userID)
			}
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := auth.GenerateToken("user-1", "correct-secret")
	if err != nil {
		t.Fatal(err)
	}

	_, err = auth.ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error with wrong secret, got nil")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	cases := []string{
		"",
		"notavalidtoken",
		"only.two",
		"a.b.c.d",
	}
	for _, tok := range cases {
		_, err := auth.ValidateToken(tok, "secret")
		if err == nil {
			t.Errorf("expected error for token %q, got nil", tok)
		}
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	password := "SuperSecret123!"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if !auth.CheckPassword(password, hash) {
		t.Error("CheckPassword() returned false for correct password")
	}

	if auth.CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword() returned true for wrong password")
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	h1, _ := auth.HashPassword("same-password")
	h2, _ := auth.HashPassword("same-password")
	if h1 == h2 {
		t.Error("two hashes of the same password should differ (random salt)")
	}
}
