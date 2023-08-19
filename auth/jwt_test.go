package auth

import (
	"crypto/rand"
	"testing"
)

func TestJwtTokens(t *testing.T) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		t.Error(err)
	}

	token, refreshToken, err := GenerateTokens(&JwtPayload{
		Username: "test",
		Admin:    true,
		UserId:   1,
	})
	if err != nil {
		t.Error(err)
	}

	payload, err := VerifyJwtToken(token)
	if err != nil {
		t.Error(err)
		return
	}

	if payload == nil {
		t.Error("payload is nil")
		return
	}
	if payload.Username != "test" {
		t.Error("username doesn't match")
	}
	if !payload.Admin {
		t.Error("admin doesn't match")
	}
	if payload.UserId != 1 {
		t.Error("userId doesn't match")
	}

	refreshPayload, err := VerifyRefreshToken(refreshToken)
	if err != nil {
		t.Error(err)
		return
	}
	if refreshPayload == nil {
		t.Error("refresh payload is nil")
		return
	}
	if refreshPayload.JwtToken != token {
		t.Error("jwtToken doesn't match")
	}
}
