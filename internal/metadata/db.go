package metadata

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/libsqladmin"
	"snitch/snitchbe/internal/metadata/sqlc"
	"snitch/snitchbe/pkg/ctxutil"

	"github.com/google/uuid"
	"github.com/tursodatabase/go-libsql"
	_ "github.com/tursodatabase/go-libsql"
)

func NewMetadataDB(ctx context.Context, token string, config dbconfig.LibSQLConfig) (*sql.DB, error) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	exists, err := libsqladmin.DoesNamespaceExist("metadata", ctx, token, config)
	if err != nil {
		slogger.ErrorContext(ctx, "Failed checking if namespace exists", "Error", err)
		return nil, fmt.Errorf("couldnt check if namespace exists: %w", err)
	}

	if !exists {
		if err := libsqladmin.CreateNamespace("metadata", ctx, token, config); err != nil {
			slogger.ErrorContext(ctx, "Failed creating metadata namespace", "Error", err)
			return nil, fmt.Errorf("couldnt create namespace: %w", err)
		}
	}

	metadataURL, err := config.MetadataDB()
	if err != nil {
		return nil, fmt.Errorf("Cant get metadata: %w", err)
	}

	dbPath := filepath.Join("/localdb", "local.metadata.db")

	conn, err := libsql.NewEmbeddedReplicaConnector(dbPath, metadataURL.String(), libsql.WithAuthToken(token))
	if err != nil {
		slogger.ErrorContext(ctx, "Error opening DB", "Error", err)
		return nil, fmt.Errorf("couldnt open db: %w", err)
	}

	db := sql.OpenDB(conn)

	queries := sqlc.New(db)

	// TODO: use transactions

	if err := queries.CreateGroupTable(ctx); err != nil {
		slogger.ErrorContext(ctx, "Failed creating group table", "Error", err)
		return nil, fmt.Errorf("couldnt create group table: %w", err)
	}

	if err := queries.CreateServerTable(ctx); err != nil {
		slogger.ErrorContext(ctx, "Failed creating server table", "Error", err)
		return nil, fmt.Errorf("couldnt create server table: %w", err)
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
