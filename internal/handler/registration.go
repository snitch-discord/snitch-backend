package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"snitch/snitchbe/internal/dbconfig"
	groupSQL "snitch/snitchbe/internal/group/sql"
	groupSQLc "snitch/snitchbe/internal/group/sqlc"
	"snitch/snitchbe/internal/libsqladmin"
	"strconv"

	"snitch/snitchbe/internal/jwt"
	metadataSQLc "snitch/snitchbe/internal/metadata/sqlc"
	"snitch/snitchbe/pkg/ctxutil"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/go-libsql"
)

type registrationRequest struct {
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
	return func(w http.ResponseWriter, r *http.Request) {
		slogger, ok := ctxutil.Value[*slog.Logger](r.Context())
		if !ok {
			slogger = slog.Default()
		}

		dbURL, err := libSqlConfig.HttpURL()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

			queries := metadataSQLc.New(db)
			var groupID uuid.UUID

			if registrationRequest.GroupID != "" {
				// Join group flow
				groupID, err = uuid.Parse(registrationRequest.GroupID)
				if err != nil {
					http.Error(w, "Invalid group ID format", http.StatusBadRequest)
					return
				}

				exists, err := libsqladmin.DoesNamespaceExist(groupID.String(), r.Context(), tokenCache.Get(), libSqlConfig)
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed checking if namespace exists", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !exists {
					http.Error(w, "Group does not exist", http.StatusNotFound)
					return
				}
				
				query := dbURL.Query()
				query.Add("authToken", tokenCache.Get())
				dbURL.RawQuery = query.Encode()
				dbURL.Host = fmt.Sprintf("%s.%s", groupID.String(), libSqlConfig.Host)

				newDB, err := sql.Open("libsql", dbURL.String())
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed to connect to db", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer newDB.Close()

				groupQueries := groupSQLc.New(newDB)
				if err := queries.AddServerToGroup(r.Context(), metadataSQLc.AddServerToGroupParams{
					GroupID:  groupID,
					ServerID: serverID,
				}); err != nil {
					slogger.ErrorContext(r.Context(), "Failed adding server to group metadata", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if err := groupQueries.AddServer(r.Context(), serverID); err != nil {
					slogger.ErrorContext(r.Context(), "Failed adding server to group", "Error", err)
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
				exists, err := libsqladmin.DoesNamespaceExist(groupID.String(), r.Context(), tokenCache.Get(), libSqlConfig)
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed checking if namespace exists", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if !exists {
					if err := libsqladmin.CreateNamespace(groupID.String(), r.Context(), tokenCache.Get(), libSqlConfig); err != nil {
						slogger.ErrorContext(r.Context(), "Failed creating namespace", "Error", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				query := dbURL.Query()
				query.Add("authToken", tokenCache.Get())
				dbURL.RawQuery = query.Encode()
				dbURL.Host = fmt.Sprintf("%s.%s", groupID.String(), libSqlConfig.Host)

				newDB, err := sql.Open("libsql", dbURL.String())
				if err != nil {
					slogger.ErrorContext(r.Context(), "Failed to connect to db", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer newDB.Close()

				groupQueries := groupSQLc.New(newDB)

				if err := newDB.PingContext(r.Context()); err != nil {
					slogger.Error("Ping Database", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if _, err := newDB.ExecContext(r.Context(), groupSQL.GroupSchema); err != nil {
					slogger.Error("Create Table", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := queries.InsertGroup(r.Context(), metadataSQLc.InsertGroupParams{
					GroupID:   groupID,
					GroupName: registrationRequest.GroupName,
				}); err != nil {
					slogger.ErrorContext(r.Context(), "Insert Group to Metadata", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := queries.AddServerToGroup(r.Context(), metadataSQLc.AddServerToGroupParams{
					GroupID:  groupID,
					ServerID: serverID,
				}); err != nil {
					slogger.ErrorContext(r.Context(), "Failed adding server to group metadata", "Error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := groupQueries.AddServer(r.Context(), serverID); err != nil {
					slogger.ErrorContext(r.Context(), "Failed adding server to group", "Error", err)
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
