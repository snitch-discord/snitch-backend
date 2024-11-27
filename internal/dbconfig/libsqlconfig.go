package dbconfig

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
)

type LibSQLConfig struct {
	Host, Port, AuthKey string
}

func LibSQLConfigFromEnv() (LibSQLConfig, error) {
	var missing []string

	get := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			missing = append(missing, key)
		}
		return val
	}

	cfg := LibSQLConfig{
		Host: get("LIBSQL_HOST"),
		Port: get("LIBSQL_PORT"),
		AuthKey: get("LIBSQL_AUTH_KEY"),
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return cfg, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

func (libsqlConfig LibSQLConfig) HttpURL() string {
	return fmt.Sprintf("http://%s:%s", libsqlConfig.Host, libsqlConfig.Port)
}

func (libsqlConfig LibSQLConfig) DatabaseURL(key ed25519.PrivateKey) string {
	return fmt.Sprintf("libsql://%s:%s?authToken=%s", libsqlConfig.Host, libsqlConfig.Port, hex.EncodeToString(key))
}
