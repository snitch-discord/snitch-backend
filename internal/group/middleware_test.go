package group_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/group"
	"snitch/snitchbe/internal/jwt"
	"testing"
)

func TestDBMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		setupDB        func(*sql.DB)
		expectedStatus int
	}{
		{
			name:           "missing server ID",
			serverID:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid server ID",
			serverID:       "not-a-number",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "server not found",
			serverID:       "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("libsql", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			if tt.setupDB != nil {
				tt.setupDB(db)
			}

			config := dbconfig.LibSQLConfig{
				Host:      "localhost",
				Port:      "8080",
				AdminPort: "9090",
				AuthKey:   "test-key",
			}

			middleware := group.NewDBMiddleware(db, config, &jwt.TokenCache{})
			handler := middleware.Handler(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.serverID != "" {
				req.Header.Set("X-Server-ID", tt.serverID)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}
		})
	}
}
