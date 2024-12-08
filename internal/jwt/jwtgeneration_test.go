package jwt_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"snitch/snitchbe/internal/jwt"
	"testing"
	"time"
)

var publicKey, privateKey, _ = ed25519.GenerateKey(rand.Reader)

func TestTokenAvailability(t *testing.T) {
	jwtDuration := 10 * time.Minute
	jwtCache := &jwt.TokenCache{}
	jwt.StartJwtGeneration(jwtDuration, jwtCache, privateKey)

	if jwtCache.Get() == "" {
		t.Error("Token cache empty")
	}
}
