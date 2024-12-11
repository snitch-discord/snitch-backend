package handler

import (
	"encoding/json"
	"net/http"
	"snitch/snitchbe/internal/group"
	groupDB "snitch/snitchbe/internal/group/db"
)

type Report struct {
	Text           string `json:"reportText"`
	ReporterID     int    `json:"reporterId,string"`     // we need to tell go that our number is encoded as a string, hence ',string'
	ReportedUserID int    `json:"reporteduserId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	ServerID       int    `json:"serverId,string"`       // we need to tell go that our number is encoded as a string, hence ',string'
}

func CreateReportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			db, err := group.GetDB(r.Context())
			if err != nil {
				http.Error(w, "Database not available", http.StatusInternalServerError)
				return
			}

			queries := groupDB.New(db)

			reports, err := queries.GetAllReports(r.Context())

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := json.NewEncoder(w).Encode(reports); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "POST":
			// w.Header().Set("Content-Type", "application/json")
			// var tournament Tournament

			// jsonerr := json.NewDecoder(r.Body).Decode(&tournament)
			// defer r.Body.Close()
			// if (jsonerr != nil) {
			// 	http.Error(w, jsonerr.Error(), http.StatusBadRequest)
			// 	return
			// }

			// statement := "INSERT INTO competitions (name, type, max_participants, random_seeds) VALUES ($1, $2, $3, $4)"
			// _, dberr := db.ExecContext(r.Context(), statement, tournament.Name, tournament.Type, tournament.MaxParticipants, tournament.RandomSeeds)
			// if (dberr != nil) {
			// 	http.Error(w, dberr.Error(), http.StatusInternalServerError)
			// 	return
			// }

			// json.NewEncoder(w).Encode(tournament)

		default:
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
