package metadata

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"snitch/snitchbe/assets"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/internal/libsqladmin"
	"snitch/snitchbe/pkg/ctxutil"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

func NewMetadataDB(ctx context.Context, tokenCache *jwt.TokenCache, config dbconfig.LibSQLConfig) (*sql.DB, error) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	exists, err := libsqladmin.DoesNamespaceExist(ctx, tokenCache, config)
	if err != nil {
		slogger.ErrorContext(ctx, "Failed checking if namespace exists", "Error", err)
		return nil, fmt.Errorf("couldnt check if namespace exists: %w", err)
	}

	if !exists {
		if err := libsqladmin.CreateNamespace(ctx, tokenCache, config); err != nil {
			slogger.ErrorContext(ctx, "Failed creating metadata namespace", "Error", err)
			return nil, fmt.Errorf("couldnt create namespace: %w", err)
		}
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
		slogger.ErrorContext(ctx, "Failed creating metadata connector", "Error", err)
		return nil, fmt.Errorf("couldnt create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	if _, err := db.ExecContext(ctx, assets.LocalDDL); err != nil {
		db.Close()
		slogger.ErrorContext(ctx, "Failed creating metadata database", "Error", err)
		return nil, fmt.Errorf("couldnt create database: %w", err)
	}

	return db, nil
}
