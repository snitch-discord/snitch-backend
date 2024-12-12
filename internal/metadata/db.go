package metadata

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/internal/libsqladmin"
	sqlc "snitch/snitchbe/internal/metadata/db"
	metadataSQL "snitch/snitchbe/internal/metadata/sql"

	"snitch/snitchbe/pkg/ctxutil"

	"github.com/google/uuid"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

func NewMetadataDB(ctx context.Context, tokenCache *jwt.TokenCache, config dbconfig.LibSQLConfig) (*sql.DB, error) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	exists, err := libsqladmin.DoesNamespaceExist("metadata", ctx, tokenCache, config)
	if err != nil {
		slogger.ErrorContext(ctx, "Failed checking if namespace exists", "Error", err)
		return nil, fmt.Errorf("couldnt check if namespace exists: %w", err)
	}

	if !exists {
		if err := libsqladmin.CreateNamespace("metadata", ctx, tokenCache, config); err != nil {
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
	if _, err := db.ExecContext(ctx, metadataSQL.MetadataSchema); err != nil {
		db.Close()
		slogger.ErrorContext(ctx, "Failed creating metadata database", "Error", err)
		return nil, fmt.Errorf("couldnt create database: %w", err)
	}

	return db, nil
}

func FindGroupIDByServerID(ctx context.Context, db *sql.DB, serverID int) (uuid.UUID, error) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	var groupID uuid.UUID
	queries := sqlc.New(db)
	groupID, err := queries.FindGroupIDByServerID(ctx, serverID)
	if err != nil {
		slogger.ErrorContext(ctx, "Failed finding group id", "Error", err)
		return uuid.Nil, fmt.Errorf("couldnt find group id: %w", err)
	}

	return groupID, nil
}

func AddServerToGroup(ctx context.Context, db *sql.DB, serverID int, groupID uuid.UUID) error {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	queries := sqlc.New(db)
	if err := queries.AddServerToGroup(ctx, sqlc.AddServerToGroupParams{
		GroupID:  groupID,
		ServerID: serverID,
	}); err != nil {
		slogger.ErrorContext(ctx, "Failed adding server to group", "Error", err)
		return fmt.Errorf("couldnt add server to group: %w", err)
	}

	return nil
}
