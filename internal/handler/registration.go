package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"snitch/snitchbe/internal/dbconfig"
	groupSQL "snitch/snitchbe/internal/group/sql"

	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/pkg/ctxutil"
	"time"

	metadataDB "snitch/snitchbe/internal/metadata/db"

	"github.com/google/uuid"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

type registrationRequest struct {
	ServerID int `json:"serverId,string"`
	UserID   int `json:"userId,string"`
}

type registrationResponse struct {
	ServerID int    `json:"serverId,string"`
	GroupID  string `json:"groupId"`
}

func CreateRegistrationHandler(tokenCache *jwt.TokenCache, db *sql.DB, libSqlConfig dbconfig.LibSQLConfig) http.HandlerFunc {
	libSQLAdminURL, err := libSqlConfig.AdminURL()
	if err != nil {
		panic(err)
	}

	libSQLHttpURL, err := libSqlConfig.HttpURL()
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		slogger, ok := ctxutil.Value[*slog.Logger](r.Context())
		if !ok {
			slogger = slog.Default()
		}

		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			var registrationRequest registrationRequest

			err := json.NewDecoder(r.Body).Decode(&registrationRequest)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			groupID := uuid.New()
			requestURL := libSQLAdminURL.JoinPath(fmt.Sprintf("v1/namespaces/%s/create", groupID))

			requestStruct := struct {
				DumpURL *string `json:"dump_url"`
			}{DumpURL: nil}

			requestBody, err := json.Marshal(requestStruct)
			if err != nil {
				slog.ErrorContext(r.Context(), "JSON Marshalling", "Error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			request, err := http.NewRequestWithContext(r.Context(), "POST", requestURL.String(), bytes.NewBuffer(requestBody))
			if err != nil {
				slogger.Error("Request Creation", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			request.Header.Add("Authorization", "Bearer "+tokenCache.Get())
			request.Header.Add("Content-Type", "application/json")
			response, err := httpClient.Do(request)
			if err != nil {
				slogger.Error("Client Call", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if response.StatusCode >= 300 || response.StatusCode < 200 {
				body, _ := io.ReadAll(response.Body)
				defer response.Body.Close()
				slogger.Error("Unexpected Response", "Status", response.Status, "StatusCode", response.StatusCode, "Body", string(body))
				http.Error(w, "Unexpected Response, Status: "+response.Status, response.StatusCode)
				return
			}

			conn, err := libsql.NewConnector(fmt.Sprintf("http://%s.%s", groupID.String(), "db"), libsql.WithProxy(libSQLHttpURL.String()), libsql.WithAuthToken(tokenCache.Get()))

			if err != nil {
				panic(err)
			}
			newDb := sql.OpenDB(conn)

			defer newDb.Close()

			if err := newDb.PingContext(r.Context()); err != nil {
				slogger.Error("Ping Database", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			result, err := newDb.ExecContext(r.Context(), groupSQL.GroupSchema)
			if err != nil {
				slogger.Error("Create Table", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			slogger.InfoContext(r.Context(), "Create Table Result", "Result", result)
			queries := metadataDB.New(db)
			if err := queries.InsertGroup(r.Context(), metadataDB.InsertGroupParams{
				GroupID:   groupID,
				GroupName: "we need the name lol",
			}); err != nil {
				slogger.ErrorContext(r.Context(), "Insert Group to Metadata", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			slogger.InfoContext(r.Context(), "Added Group to Metadata", "Result", groupID.String())

			if err = json.NewEncoder(w).Encode(registrationResponse{ServerID: registrationRequest.ServerID, GroupID: groupID.String()}); err != nil {
				slogger.Error("Encode Response", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
