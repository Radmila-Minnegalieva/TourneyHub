package security

import (
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSecretTokens(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		token, err := NewToken()
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil || len(decoded) != 32 {
			t.Fatalf("token must encode 256 random bits: %q (%v)", token, err)
		}
		if seen[token] {
			t.Fatal("duplicate token")
		}
		seen[token] = true
		if hash := TokenHash(token); len(hash) != 64 || hash == token || hash != TokenHash(token) {
			t.Fatalf("invalid stable SHA-256 hash: %q", hash)
		}
	}
}

func TestPasswordBoundsAndHash(t *testing.T) {
	for _, tc := range []struct {
		name     string
		password string
		valid    bool
	}{
		{"short", "short", false},
		{"unicode below minimum characters", "пароль", false},
		{"minimum", "twelve-chars!", true},
		{"unicode", "парольдлинный", true},
		{"maximum", strings.Repeat("a", 72), true},
		{"too many bytes", strings.Repeat("a", 73), false},
		{"unicode exceeds bcrypt bytes", strings.Repeat("я", 37), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			if (err == nil) != tc.valid {
				t.Fatalf("hashPassword error=%v, valid=%v", err, tc.valid)
			}
			if !tc.valid {
				return
			}
			if hash == tc.password || bcrypt.CompareHashAndPassword([]byte(hash), []byte(tc.password)) != nil {
				t.Fatal("password was not stored as a usable bcrypt hash")
			}
			if bcrypt.CompareHashAndPassword([]byte(hash), []byte("x"+tc.password[1:])) == nil {
				t.Fatal("wrong password accepted")
			}
		})
	}
}
