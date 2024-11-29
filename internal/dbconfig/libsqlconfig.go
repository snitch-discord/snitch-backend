package dbconfig

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/url"
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

func (libSQLConfig LibSQLConfig) HttpURL() (*url.URL, error) {
	return url.Parse(fmt.Sprintf("http://%s:%s", libSQLConfig.Host, libSQLConfig.Port))
}

func (libSQLConfig LibSQLConfig) DatabaseURL(key ed25519.PrivateKey) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("libsql://%s:%s?authToken=%s", libSQLConfig.Host, libSQLConfig.Port, hex.EncodeToString(key)))
}
