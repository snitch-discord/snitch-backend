package metadata

import (
	"context"
	"database/sql"
	"fmt"
	"snitch/snitchbe/assets"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/internal/libsqladmin"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

func NewMetadataDB(ctx context.Context, tokenCache *jwt.TokenCache, config dbconfig.LibSQLConfig) (*sql.DB, error) {
	if err := libsqladmin.CreateNamespace(ctx, tokenCache, config); err != nil {
		return nil, fmt.Errorf("create namespace: %w", err)
	}

	httpURL, err := config.HttpURL()
	if err != nil {
		return nil, fmt.Errorf("get http url: %w", err)
	}

	connector, err := libsql.NewConnector(
		fmt.Sprintf("http://%s.%s", "metadata", "db"),
		libsql.WithProxy(httpURL.String()),
		libsql.WithAuthToken(tokenCache.Get()),
	)
	if err != nil {
		return nil, fmt.Errorf("create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	if _, err := db.ExecContext(ctx, assets.LocalDDL); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return db, nil
}
