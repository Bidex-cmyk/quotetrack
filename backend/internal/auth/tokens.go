package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	hashCost = 12
)

// Password hashing (bcrypt). Costs are configurable-ish via hashCost const.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), hashCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// TokenManager signs and verifies JWT bearer tokens.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

type claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

// Create signs a token for userID.
func (tm *TokenManager) Create(userID string) (string, error) {
	now := time.Now()
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.ttl)),
			Issuer:    "quotetrack",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	s, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return s, nil
}

// Verify parses and validates a token, returning the user ID.
func (tm *TokenManager) Verify(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return tm.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	if c.UserID == "" {
		return "", fmt.Errorf("token missing user")
	}
	return c.UserID, nil
}

// RandomPublicID generates a secure random identifier for share links.
func RandomPublicID() (string, error) {
	b := make([]byte, 18) // 144 bits of entropy
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
