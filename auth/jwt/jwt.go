package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ao9911/go-matrix/util/xtime"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	jtiBytes         = 16
)

// Config defines access and refresh token signing options.
type Config struct {
	Issuer        string         `json:"issuer" toml:"issuer"`
	AccessSecret  string         `json:"access_secret" toml:"access_secret"`
	RefreshSecret string         `json:"refresh_secret" toml:"refresh_secret"`
	AccessExpire  xtime.Duration `json:"access_expire" toml:"access_expire"`
	RefreshExpire xtime.Duration `json:"refresh_expire" toml:"refresh_expire"`
}

// TokenPair is the generated access/refresh token response.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims carries the token type and standard registered claims.
type Claims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWT generates and parses access/refresh JWT tokens.
type JWT struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
}

// NewJWT creates a JWT helper from validated config.
func NewJWT(c *Config) *JWT {
	if c == nil {
		panic("jwt: nil config")
	}
	if c.AccessSecret == "" {
		panic("jwt: empty access secret")
	}
	if c.RefreshSecret == "" {
		panic("jwt: empty refresh secret")
	}
	if c.AccessExpire <= 0 {
		panic("jwt: invalid access expire")
	}
	if c.RefreshExpire <= 0 {
		panic("jwt: invalid refresh expire")
	}

	return &JWT{
		accessSecret:  []byte(c.AccessSecret),
		refreshSecret: []byte(c.RefreshSecret),
		accessExpire:  time.Duration(c.AccessExpire),
		refreshExpire: time.Duration(c.RefreshExpire),
		issuer:        c.Issuer,
	}
}

// GeneratePair creates access and refresh tokens for the subject.
func (j *JWT) GeneratePair(subject string) (*TokenPair, error) {
	if subject == "" {
		return nil, errors.New("jwt: empty subject")
	}

	now := time.Now()
	accessToken, err := j.generateToken(subject, TokenTypeAccess, j.accessExpire, j.accessSecret, now)
	if err != nil {
		return nil, err
	}
	refreshToken, err := j.generateToken(subject, TokenTypeRefresh, j.refreshExpire, j.refreshSecret, now)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ParseAccess parses and validates an access token.
func (j *JWT) ParseAccess(tokenString string) (*Claims, error) {
	return j.parseToken(tokenString, TokenTypeAccess, j.accessSecret)
}

// ParseRefresh parses and validates a refresh token.
func (j *JWT) ParseRefresh(tokenString string) (*Claims, error) {
	return j.parseToken(tokenString, TokenTypeRefresh, j.refreshSecret)
}

func newJTI() (string, error) {
	buf := make([]byte, jtiBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (j *JWT) generateToken(subject, tokenType string, expire time.Duration, secret []byte, now time.Time) (string, error) {
	jti, err := newJTI()
	if err != nil {
		return "", fmt.Errorf("jwt: generate %s jti: %w", tokenType, err)
	}

	expiresAt := jwt.NewNumericDate(now.Add(expire))
	claims := &Claims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   subject,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: expiresAt,
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("jwt: sign %s token: %w", tokenType, err)
	}
	return token, nil
}

func (j *JWT) parseToken(tokenString, tokenType string, secret []byte) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("jwt: empty token")
	}

	claims := &Claims{}
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	}
	if j.issuer != "" {
		opts = append(opts, jwt.WithIssuer(j.issuer))
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("jwt: invalid signing method: %v", token.Header["alg"])
		}
		return secret, nil
	}, opts...)
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	if claims.Subject == "" {
		return nil, errors.New("jwt: empty subject")
	}
	if claims.TokenType != tokenType {
		return nil, fmt.Errorf("jwt: invalid token type: %s", claims.TokenType)
	}
	return claims, nil
}
