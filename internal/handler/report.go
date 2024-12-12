package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"snitch/snitchbe/internal/group"
	groupDB "snitch/snitchbe/internal/group/db"
	"snitch/snitchbe/pkg/ctxutil"
)

type reportRequest struct {
	ReportText     string `json:"reportText"`
	ReporterID     int    `json:"reporterId,string"`
	ReportedUserID int    `json:"reportedUserId,string"`
}

func CreateReportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slogger, ok := ctxutil.Value[*slog.Logger](r.Context())
		if !ok {
			slogger = slog.Default()
		}

		switch r.Method {
		case "GET":
			db, err := group.GetDB(r.Context())
			if err != nil {
				slogger.Error("failed to get db", "Error", err)
				http.Error(w, "Database not available", http.StatusInternalServerError)
				return
			}

			queries := groupDB.New(db)

			reports, err := queries.GetAllReports(r.Context())

			if err != nil {
				slogger.Error("failed to get reports", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := json.NewEncoder(w).Encode(reports); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "POST":

			var reportRequest reportRequest
			err := json.NewDecoder(r.Body).Decode(&reportRequest)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			db, err := group.GetDB(r.Context())

			if err != nil {
				slogger.Error("failed to get db", "Error", err)
				http.Error(w, "Database not available", http.StatusInternalServerError)
				return
			}
			queries := groupDB.New(db)

			if err := queries.AddUser(r.Context(), reportRequest.ReportedUserID); err != nil {
				slogger.Error(fmt.Sprintf("failed to add user %d", reportRequest.ReportedUserID), "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := queries.AddUser(r.Context(), reportRequest.ReporterID); err != nil {
				slogger.Error(fmt.Sprintf("failed to add user %d", reportRequest.ReporterID), "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			serverID, err := group.GetServerID(r.Context())
			if err != nil {
				slogger.Error("failed to get server id", "Error", err)
				http.Error(w, "Server ID not available", http.StatusInternalServerError)
				return
			}

			reportID, err := queries.CreateReport(r.Context(), groupDB.CreateReportParams{
				OriginServerID: serverID,
				ReportText:     reportRequest.ReportText,
				ReporterID:     reportRequest.ReporterID,
				ReportedUserID: reportRequest.ReportedUserID,
			})

			if err != nil {
				slogger.Error("failed to create report", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": reportID}); err != nil {
				slogger.Error("failed to encode response", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		default:
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
