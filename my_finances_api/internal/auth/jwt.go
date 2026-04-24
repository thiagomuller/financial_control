package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtClaims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

// GenerateToken creates a signed HS256 JWT containing the user ID.
func GenerateToken(userID, secret string) (string, error) {
	headerJSON, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	now := time.Now()
	claimsJSON, err := json.Marshal(jwtClaims{
		Sub: userID,
		Exp: now.Add(24 * time.Hour).Unix(),
		Iat: now.Unix(),
	})
	if err != nil {
		return "", err
	}

	h := base64.RawURLEncoding.EncodeToString(headerJSON)
	p := base64.RawURLEncoding.EncodeToString(claimsJSON)
	sig := signHS256(h+"."+p, secret)

	return h + "." + p + "." + sig, nil
}

// ValidateToken verifies the token signature and expiry, returning the user ID.
func ValidateToken(token, secret string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	expected := signHS256(parts[0]+"."+parts[1], secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", fmt.Errorf("invalid signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decoding payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return "", fmt.Errorf("parsing claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return "", fmt.Errorf("token expired")
	}

	return claims.Sub, nil
}

func signHS256(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// HashPassword produces a salted SHA-256 hash for a password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	saltHex := hex.EncodeToString(salt)
	h := sha256.New()
	h.Write([]byte(saltHex + password))
	return saltHex + ":" + hex.EncodeToString(h.Sum(nil)), nil
}

// CheckPassword validates a plaintext password against a stored hash.
func CheckPassword(password, stored string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	h := sha256.New()
	h.Write([]byte(parts[0] + password))
	actual := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(actual), []byte(parts[1]))
}
