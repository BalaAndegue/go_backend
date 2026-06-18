package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cr3t-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "s3cr3t-password" {
		t.Fatal("password was not hashed")
	}
	if !CheckPassword("s3cr3t-password", hash) {
		t.Error("CheckPassword rejected the correct password")
	}
	if CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword accepted an incorrect password")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken(42, "ADMIN", 3)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if got := uint(claims["user_id"].(float64)); got != 42 {
		t.Errorf("user_id = %d, want 42", got)
	}
	if got := claims["role"].(string); got != "ADMIN" {
		t.Errorf("role = %q, want ADMIN", got)
	}
	if got := int(claims["ver"].(float64)); got != 3 {
		t.Errorf("ver = %d, want 3", got)
	}
	if got := claims["type"].(string); got != "access" {
		t.Errorf("type = %q, want access", got)
	}
}

func TestRefreshTokenClaims(t *testing.T) {
	token, err := GenerateRefreshToken(7, 1)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if got := claims["type"].(string); got != "refresh" {
		t.Errorf("type = %q, want refresh", got)
	}
}

func TestValidateTokenRejectsGarbage(t *testing.T) {
	if _, err := ValidateToken("not-a-real-token"); err == nil {
		t.Error("ValidateToken accepted a malformed token")
	}
}

func TestGenerateSlug(t *testing.T) {
	cases := map[string]string{
		"Hello World":       "hello-world",
		"  Trim  Me  ":      "trim-me",
		"Accents & Symbols": "accents-symbols",
		"Already-slug":      "already-slug",
	}
	for in, want := range cases {
		if got := GenerateSlug(in); got != want {
			t.Errorf("GenerateSlug(%q) = %q, want %q", in, got, want)
		}
	}
}
