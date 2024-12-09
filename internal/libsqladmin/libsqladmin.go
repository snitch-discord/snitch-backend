package libsqladmin

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/pkg/ctxutil"
)

func CreateNamespace(ctx context.Context, tokenCache *jwt.TokenCache, config dbconfig.LibSQLConfig) error {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	adminURL, err := config.AdminURL()
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "POST",
		adminURL.JoinPath("v1/namespaces/metadata/create").String(),
		bytes.NewReader([]byte(`{"dump_url": null}`)))

	if err != nil {
		slogger.ErrorContext(ctx, "Failed creating metadata namespace create request", "Error", err)
		return err
	}

	request.Header.Set("Authorization", "Bearer "+tokenCache.Get())
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		slogger.ErrorContext(ctx, "Failed executing metadata namespace create request", "Error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusConflict {
		slogger.ErrorContext(ctx, "Unexpected status code", "Status", resp.Status)
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}
