package jwt

import (
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
)

func TestJWT_GeneratePairParseAccessParseRefresh(t *testing.T) {
	j, err := NewJWT(&Config{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessExpire:  xtime.Duration(time.Minute),
		RefreshExpire: xtime.Duration(time.Hour),
		Issuer:        "go-matrix-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	pair, err := j.GeneratePair("user-123")
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("expected different access and refresh tokens")
	}

	accessClaims, err := j.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	refreshClaims, err := j.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}

	if accessClaims.Subject != "user-123" || refreshClaims.Subject != "user-123" {
		t.Fatalf("unexpected subjects: access=%q refresh=%q", accessClaims.Subject, refreshClaims.Subject)
	}
	if accessClaims.TokenType != TokenTypeAccess || refreshClaims.TokenType != TokenTypeRefresh {
		t.Fatalf("unexpected token types: access=%q refresh=%q", accessClaims.TokenType, refreshClaims.TokenType)
	}
	if accessClaims.Issuer != "go-matrix-test" || refreshClaims.Issuer != "go-matrix-test" {
		t.Fatalf("unexpected issuers: access=%q refresh=%q", accessClaims.Issuer, refreshClaims.Issuer)
	}
	if accessClaims.ExpiresAt == nil || refreshClaims.ExpiresAt == nil {
		t.Fatal("expected access and refresh expirations")
	}
	if !refreshClaims.ExpiresAt.Time.After(accessClaims.ExpiresAt.Time) {
		t.Fatalf("expected refresh to expire after access: refresh=%s access=%s", refreshClaims.ExpiresAt.Time, accessClaims.ExpiresAt.Time)
	}

	for _, id := range []string{accessClaims.ID, refreshClaims.ID} {
		if len(id) != jtiBytes*2 {
			t.Fatalf("expected %d-char jti, got %q", jtiBytes*2, id)
		}
		for _, r := range id {
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
				t.Fatalf("expected lowercase hex jti, got %q", id)
			}
		}
	}
	if accessClaims.ID == refreshClaims.ID {
		t.Fatal("expected different access and refresh jti")
	}
}
