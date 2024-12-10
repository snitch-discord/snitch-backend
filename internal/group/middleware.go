package group

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/internal/metadata"
	"strconv"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

type DatabaseContext string

const (
	ServerIDHeader = "X-Server-ID"
	DBContextKey   = DatabaseContext("db")
)

type DBMiddleware struct {
	metadataDB *sql.DB
	config     dbconfig.LibSQLConfig
	tokenCache *jwt.TokenCache
}

func NewDBMiddleware(metadataDB *sql.DB, config dbconfig.LibSQLConfig, tokenCache *jwt.TokenCache) *DBMiddleware {
	return &DBMiddleware{
		metadataDB: metadataDB,
		config:     config,
		tokenCache: tokenCache,
	}
}

func (m *DBMiddleware) getServerID(r *http.Request) (int, error) {
	serverIDStr := r.Header.Get(ServerIDHeader)
	if serverIDStr == "" {
		return 0, fmt.Errorf("server ID header is required")
	}

	serverID, err := strconv.Atoi(serverIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid server ID format")
	}

	return serverID, nil
}

func (m *DBMiddleware) connectToDB(ctx context.Context, groupID string) (*sql.DB, error) {
	httpURL, err := m.config.HttpURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get HTTP URL: %w", err)
	}

	connector, err := libsql.NewConnector(
		fmt.Sprintf("http://%s.%s", groupID, "db"),
		libsql.WithProxy(httpURL.String()),
		libsql.WithAuthToken(m.tokenCache.Get()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	return sql.OpenDB(connector), nil
}

func (m *DBMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID, err := m.getServerID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		groupID, err := metadata.FindGroupIDByServerID(r.Context(), m.metadataDB, serverID)
		if err != nil {
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		db, err := m.connectToDB(r.Context(), groupID.String())
		if err != nil {
			http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		ctx := context.WithValue(r.Context(), DBContextKey, db)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetDB(ctx context.Context) (*sql.DB, error) {
	db, ok := ctx.Value(DBContextKey).(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database not found in context")
	}
	return db, nil
}
