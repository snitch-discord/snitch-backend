package libsqladmin

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
)

func CreateNamespace(ctx context.Context, tokenCache *jwt.TokenCache, config dbconfig.LibSQLConfig) error {
	adminURL, err := config.AdminURL()
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "POST",
		adminURL.JoinPath("v1/namespaces/metadata/create").String(),
		bytes.NewReader([]byte(`{"dump_url": null}`)))

	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+tokenCache.Get())
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}
