package jwt

import (
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
)

var (
	jwtConfig *Config
	jwtClient *JWT
)

func init() {
	jwtConfig = &Config{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessExpire:  xtime.Duration(time.Minute),
		RefreshExpire: xtime.Duration(time.Hour),
		Issuer:        "go-matrix-test",
	}
	jwtClient = NewJWT(jwtConfig)
}

// go test -v -test.run TestNewJWT
func TestNewJWT(t *testing.T) {
	if jwtClient == nil {
		t.Fatal("NewJWT() returned nil")
	}
	if string(jwtClient.accessSecret) != jwtConfig.AccessSecret {
		t.Fatalf("access secret = %q, want %q", string(jwtClient.accessSecret), jwtConfig.AccessSecret)
	}
	if string(jwtClient.refreshSecret) != jwtConfig.RefreshSecret {
		t.Fatalf("refresh secret = %q, want %q", string(jwtClient.refreshSecret), jwtConfig.RefreshSecret)
	}
	if jwtClient.accessExpire != time.Duration(jwtConfig.AccessExpire) {
		t.Fatalf("access expire = %s, want %s", jwtClient.accessExpire, time.Duration(jwtConfig.AccessExpire))
	}
	if jwtClient.refreshExpire != time.Duration(jwtConfig.RefreshExpire) {
		t.Fatalf("refresh expire = %s, want %s", jwtClient.refreshExpire, time.Duration(jwtConfig.RefreshExpire))
	}
	if jwtClient.issuer != jwtConfig.Issuer {
		t.Fatalf("issuer = %q, want %q", jwtClient.issuer, jwtConfig.Issuer)
	}
}

// go test -v -test.run TestJWT_GeneratePair
func TestJWT_GeneratePair(t *testing.T) {
	pair, err := jwtClient.GeneratePair("user-123")
	if err != nil {
		t.Fatal(err)
	}
	if pair == nil {
		t.Fatal("GeneratePair() returned nil")
	}
	if pair.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if pair.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("expected different access and refresh tokens")
	}

	if _, err := jwtClient.GeneratePair(""); err == nil {
		t.Fatal("expected empty subject error")
	}
}

// go test -v -test.run TestJWT_ParseAccess
func TestJWT_ParseAccess(t *testing.T) {
	pair := mustGeneratePair(t)

	claims, err := jwtClient.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	assertClaims(t, claims, "user-123", TokenTypeAccess)

	if _, err := jwtClient.ParseAccess(pair.RefreshToken); err == nil {
		t.Fatal("expected refresh token to be rejected as access token")
	}
}

// go test -v -test.run TestJWT_ParseRefresh
func TestJWT_ParseRefresh(t *testing.T) {
	pair := mustGeneratePair(t)

	claims, err := jwtClient.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	assertClaims(t, claims, "user-123", TokenTypeRefresh)

	if _, err := jwtClient.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("expected access token to be rejected as refresh token")
	}
}

func mustGeneratePair(t *testing.T) *TokenPair {
	t.Helper()

	pair, err := jwtClient.GeneratePair("user-123")
	if err != nil {
		t.Fatal(err)
	}
	if pair == nil {
		t.Fatal("GeneratePair() returned nil")
	}
	return pair
}

func assertClaims(t *testing.T, claims *Claims, subject, tokenType string) {
	t.Helper()

	if claims == nil {
		t.Fatal("expected claims")
	}
	if claims.Subject != subject {
		t.Fatalf("subject = %q, want %q", claims.Subject, subject)
	}
	if claims.TokenType != tokenType {
		t.Fatalf("token type = %q, want %q", claims.TokenType, tokenType)
	}
	if claims.Issuer != jwtConfig.Issuer {
		t.Fatalf("issuer = %q, want %q", claims.Issuer, jwtConfig.Issuer)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected expiration")
	}
	if len(claims.ID) != jtiBytes*2 {
		t.Fatalf("jti length = %d, want %d", len(claims.ID), jtiBytes*2)
	}
	for _, r := range claims.ID {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			t.Fatalf("expected lowercase hex jti, got %q", claims.ID)
		}
	}
}
