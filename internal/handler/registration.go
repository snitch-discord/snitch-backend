package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"snitch/snitchbe/internal/dbconfig"
	groupSQL "snitch/snitchbe/internal/group/sql"
	"snitch/snitchbe/internal/libsqladmin"
	"strconv"

	"snitch/snitchbe/internal/jwt"
	metadataDB "snitch/snitchbe/internal/metadata/db"
	"snitch/snitchbe/pkg/ctxutil"

	"github.com/google/uuid"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

type registrationRequest struct {
	ServerID  int    `json:"serverId,string"`
	UserID    int    `json:"userId,string"`
	GroupID   string `json:"groupId,omitempty"`
	GroupName string `json:"groupName,omitempty"`
}

type registrationResponse struct {
	ServerID int    `json:"serverId,string"`
	GroupID  string `json:"groupId"`
}

const ServerIDHeader = "X-Server-ID"

func getServerIDFromHeader(r *http.Request) (int, error) {
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
func CreateRegistrationHandler(tokenCache *jwt.TokenCache, db *sql.DB, libSqlConfig dbconfig.LibSQLConfig) http.HandlerFunc {
	libSQLHttpURL, err := libSqlConfig.HttpURL()
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		slogger, ok := ctxutil.Value[*slog.Logger](r.Context())
		if !ok {
			slogger = slog.Default()
		}

		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")

			serverID, err := getServerIDFromHeader(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var registrationRequest registrationRequest
			err = json.NewDecoder(r.Body).Decode(&registrationRequest)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			queries := metadataDB.New(db)
			var groupID uuid.UUID

			if registrationRequest.GroupID != "" {
				// Join group flow
				groupID, err = uuid.Parse(registrationRequest.GroupID)
				if err != nil {
					http.Error(w, "Invalid group ID format", http.StatusBadRequest)
					return
				}

				exists, err := libsqladmin.DoesNamespaceExist(groupID.String(), r.Context(), tokenCache, libSqlConfig)
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed checking if namespace exists", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !exists {
					http.Error(w, "Group does not exist", http.StatusNotFound)
					return
				}

				if err := queries.AddServerToGroup(r.Context(), metadataDB.AddServerToGroupParams{
					GroupID:  groupID,
					ServerID: serverID,
				}); err != nil {
					slogger.ErrorContext(r.Context(), "Failed adding server to group metadata", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

			} else {
				// Create new group flow
				if registrationRequest.GroupName == "" {
					http.Error(w, "Group name is required when creating a new group", http.StatusBadRequest)
					return
				}

				groupID = uuid.New()
				exists, err := libsqladmin.DoesNamespaceExist(groupID.String(), r.Context(), tokenCache, libSqlConfig)
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed checking if namespace exists", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if !exists {
					if err := libsqladmin.CreateNamespace(groupID.String(), r.Context(), tokenCache, libSqlConfig); err != nil {
						slogger.ErrorContext(r.Context(), "Failed creating namespace", "Error", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				conn, err := libsql.NewConnector(
					fmt.Sprintf("http://%s.%s", groupID.String(), "db"),
					libsql.WithProxy(libSQLHttpURL.String()),
					libsql.WithAuthToken(tokenCache.Get()),
				)
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed creating database connector", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				newDb := sql.OpenDB(conn)
				defer newDb.Close()

				if err := newDb.PingContext(r.Context()); err != nil {
					slogger.Error("Ping Database", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if _, err := newDb.ExecContext(r.Context(), groupSQL.GroupSchema); err != nil {
					slogger.Error("Create Table", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := queries.InsertGroup(r.Context(), metadataDB.InsertGroupParams{
					GroupID:   groupID,
					GroupName: registrationRequest.GroupName,
				}); err != nil {
					slogger.ErrorContext(r.Context(), "Insert Group to Metadata", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := queries.AddServerToGroup(r.Context(), metadataDB.AddServerToGroupParams{
					GroupID:  groupID,
					ServerID: serverID,
				}); err != nil {
					slogger.ErrorContext(r.Context(), "Failed adding server to group metadata", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			slogger.InfoContext(r.Context(), "Registration completed",
				"groupID", groupID.String(),
				"serverID", serverID,
				"isNewGroup", registrationRequest.GroupID == "")

			if err = json.NewEncoder(w).Encode(registrationResponse{
				ServerID: serverID,
				GroupID:  groupID.String(),
			}); err != nil {
				slogger.Error("Encode Response", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
