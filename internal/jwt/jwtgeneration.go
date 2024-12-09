package jwt

import (
	"crypto/ed25519"
	"log/slog"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenCache struct {
	token string
	mutex sync.Mutex
}

func (tokenCache *TokenCache) set(token string) {
	tokenCache.mutex.Lock()
	defer tokenCache.mutex.Unlock()
	tokenCache.token = token
}

func (tokenCache *TokenCache) Get() string {
	tokenCache.mutex.Lock()
	defer tokenCache.mutex.Unlock()
	return tokenCache.token
}

func createJwtGenerator(key ed25519.PrivateKey) func(time.Duration) (string, error) {
	return func(duration time.Duration) (string, error) {
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
			"exp": time.Now().Add(duration).Unix(),
		})

		return token.SignedString(key)
	}
}

func startJwtTicker(interval time.Duration, tokenCache *TokenCache, jwtGenerator func(time.Duration) (string, error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		token, err := jwtGenerator(interval)
		if err != nil {
			slog.Error("Error generating token", "Error", err)
			continue
		}
		tokenCache.set(token)
	}
}

func StartJwtGeneration(interval time.Duration, tokenCache *TokenCache, key ed25519.PrivateKey) {
	jwtGenerator := createJwtGenerator(key)
	firstToken, err := jwtGenerator(interval)
	if err != nil {
		slog.Error("Error generating token", "Error", err)
	}
	tokenCache.set(firstToken)
	
	go startJwtTicker(interval, tokenCache, jwtGenerator)
}
