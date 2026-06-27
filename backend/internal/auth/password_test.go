package auth

import "testing"

func TestHashAndComparePassword(t *testing.T) {
	password := "super-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}
	if hash == password {
		t.Fatal("hash should not equal plaintext password")
	}

	if err := ComparePassword(hash, password); err != nil {
		t.Fatalf("ComparePassword() should accept correct password, got %v", err)
	}
	if err := ComparePassword(hash, "wrong-password"); err == nil {
		t.Fatal("ComparePassword() should reject incorrect password")
	}
}
