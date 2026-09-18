package auth

import (
	"testing"
	"time"
)

func TestPasswordHashAndCheck(t *testing.T) {
	plain := "s3cret-pass"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatal(err)
	}
	if hash == plain {
		t.Error("hash must not equal plaintext")
	}
	if !CheckPassword(hash, plain) {
		t.Error("correct password should verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("wrong password should not verify")
	}
}

func TestTokenCreateVerify(t *testing.T) {
	tm := NewTokenManager("unit-test-secret", time.Hour)
	tok, err := tm.Create("user-123")
	if err != nil {
		t.Fatal(err)
	}
	uid, err := tm.Verify(tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if uid != "user-123" {
		t.Errorf("got uid %q", uid)
	}
}

func TestTokenWrongSecret(t *testing.T) {
	tm1 := NewTokenManager("secret-one", time.Hour)
	tm2 := NewTokenManager("secret-two", time.Hour)
	tok, _ := tm1.Create("user-123")
	if _, err := tm2.Verify(tok); err == nil {
		t.Error("token signed with different secret should fail")
	}
}

func TestTokenExpired(t *testing.T) {
	tm := NewTokenManager("secret", -time.Minute)
	tok, _ := tm.Create("user-123")
	if _, err := tm.Verify(tok); err == nil {
		t.Error("expired token should fail")
	}
}

func TestRandomPublicID(t *testing.T) {
	a, err := RandomPublicID()
	if err != nil {
		t.Fatal(err)
	}
	b, err := RandomPublicID()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Error("public IDs should be unique")
	}
	if len(a) < 20 {
		t.Errorf("public ID too short: %q", a)
	}
}

func TestBearerToken(t *testing.T) {
	r := newGetRequest("Bearer abc.def")
	if got := bearerToken(r); got != "abc.def" {
		t.Errorf("got %q", got)
	}
	r = newGetRequest("")
	if got := bearerToken(r); got != "" {
		t.Errorf("got %q", got)
	}
}