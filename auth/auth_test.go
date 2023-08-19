package auth

import (
	"testing"
)

func TestPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}

	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Error(err)
	}

	if !match {
		t.Error("passwords don't match")
	}
}
