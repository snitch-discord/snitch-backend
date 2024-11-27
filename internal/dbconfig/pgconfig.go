package dbconfig

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
)

type PgConfig struct {
	user, database, host, password, port string
	sslMode string // optional
}

func PgConfigFromEnv() (PgConfig, error) {
	var missing []string

	get := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			missing = append(missing, key)
		}
		return val
	}

	cfg := PgConfig{
		user: get("PG_USER"),
		database: get("PG_DATABASE"),
		host: get("PG_HOST"),
		password: get("PG_PASSWORD"),
		port: get("PG_PORT"),
		sslMode: os.Getenv("PG_SSLMODE"), // optional
	}

	switch cfg.sslMode {
	case "", "disable", "allow", "require", "verify-ca", "verify-full":
	default:
		return cfg, fmt.Errorf(`invalid sslmode %s, expected one of: "", "disable", "allow", "require", "verify-ca", or "verify-full"`, cfg.sslMode)
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return cfg, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

func (pg PgConfig) String() string {
	s := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pg.user, pg.password, pg.host, pg.port, pg.database)

	slog.Info(s)

	if pg.sslMode != "" {
		s += "?sslmode=" + pg.sslMode
	}

	return s
}
