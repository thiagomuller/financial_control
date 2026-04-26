package unit_test

import (
	"encoding/base64"
	"strings"
	"testing"
)

func iconURLExceedsLimit(iconURL string, maxBytes int) bool {
	raw := iconURL
	if idx := strings.Index(raw, ","); idx != -1 {
		raw = raw[idx+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return false
	}
	return len(decoded) > maxBytes
}

func TestIconURLExceedsLimit(t *testing.T) {
	const maxBytes = 5 * 1024 * 1024

	smallPayload := base64.StdEncoding.EncodeToString([]byte("small"))
	smallURL := "data:image/png;base64," + smallPayload

	if iconURLExceedsLimit(smallURL, maxBytes) {
		t.Error("small icon URL should not exceed limit")
	}

	bigPayload := base64.StdEncoding.EncodeToString(make([]byte, maxBytes+1))
	bigURL := "data:image/png;base64," + bigPayload

	if !iconURLExceedsLimit(bigURL, maxBytes) {
		t.Error("big icon URL should exceed limit")
	}
}

func TestIconURLWithoutDataPrefix(t *testing.T) {
	const maxBytes = 5 * 1024 * 1024

	raw := base64.StdEncoding.EncodeToString([]byte("hello"))
	if iconURLExceedsLimit(raw, maxBytes) {
		t.Error("raw base64 without data prefix should not exceed limit for small input")
	}
}

func TestUpdateBankAccountRequestAcceptsIconURL(t *testing.T) {
	iconURL := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg=="
	req := struct {
		Name    string
		IconURL *string
	}{
		Name:    "My Bank",
		IconURL: &iconURL,
	}

	if req.IconURL == nil {
		t.Fatal("IconURL should not be nil")
	}
	if *req.IconURL != iconURL {
		t.Errorf("IconURL = %q, want %q", *req.IconURL, iconURL)
	}
}

func TestUpdateBankAccountRequestIconURLNilForRemoval(t *testing.T) {
	req := struct {
		Name    string
		IconURL *string
	}{
		Name:    "My Bank",
		IconURL: nil,
	}

	if req.IconURL != nil {
		t.Error("nil IconURL should represent icon removal")
	}
}
