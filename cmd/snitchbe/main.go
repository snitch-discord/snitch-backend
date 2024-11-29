package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"snitch/snitchbe/assets"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/pkg/ctxutil"
	"snitch/snitchbe/pkg/middleware"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/go-libsql"
)

type Report struct {
	Text string `json:"reportText"`
	ReporterID int `json:"reporterId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	ReportedUserID int `json:"reporteduserId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	ServerID int `json:"serverId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
}

func createReportHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch (r.Method) {
		case "GET":
			// w.Header().Set("Content-Type", "application/json")
			// var tournaments []Tournament

			// statement := "SELECT * FROM competitions"
			// rows, err := db.QueryContext(r.Context(), statement)
			// if (err != nil) {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }

			// for rows.Next() {
			// 	var tournament Tournament
			// 	var tournamentId int

			// 	err = rows.Scan(&tournamentId, &tournament.Name, &tournament.Type, &tournament.MaxParticipants, &tournament.RandomSeeds)

			// 	if err != nil {
			// 		break
			// 	}

			// 	tournaments = append(tournaments, tournament)
			// }

			// json.NewEncoder(w).Encode(tournaments)
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

type registrationRequest struct {
	ServerID int `json:"serverId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	UserID int `json:"userId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
}

type registrationResponse struct {
	ServerID int `json:"serverId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	GroupID string `json:"groupId"`
}

func createRegistrationHandler(tokenCache *jwt.TokenCache, db *sql.DB, libSqlConfig dbconfig.LibSQLConfig, key ed25519.PrivateKey) http.HandlerFunc {
	newDatabaseURL, err := libSqlConfig.DatabaseURL(key)
	if (err != nil) {
		panic(err)
	}

	libSQLHttpURL, err := libSqlConfig.HttpURL()
	if (err != nil) {
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

		switch (r.Method) {
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
			requestURL := libSQLHttpURL.JoinPath(fmt.Sprintf("v1/namespaces/%s/create", groupID))

			request, err := http.NewRequestWithContext(r.Context(), "GET", requestURL.String(), nil)
			if err != nil {
				slogger.Error("Request Creation", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			request.Header.Add("Authorization", "Bearer " + tokenCache.Get())
			slogger.Info("DEBUG", "Header", request.Header)
			response, err := httpClient.Do(request)
			if err != nil {
				slogger.Error("Client Call", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			responseBodyString := new(strings.Builder)
			_, err = io.Copy(responseBodyString, response.Body)
			defer response.Body.Close()
			if (err != nil) {
				slogger.Error("Create String Buffer", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			slogger.InfoContext(r.Context(), "Response", "Body", responseBodyString)

			newDatabaseURL.Host = fmt.Sprintf("%s.%s", groupID.String(), newDatabaseURL.Host)
			slogger.Info("New URL", "URL", newDatabaseURL)

			newDb, err := sql.Open("libsql", newDatabaseURL.String());
			if err != nil {
				panic(err)
			}
			defer newDb.Close()

			if err := newDb.PingContext(r.Context()); err != nil {
				slogger.Error("Ping Database", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			slogger.InfoContext(r.Context(), "DDL", "Remote DDL", assets.RemoteDDL)
			result, err := db.ExecContext(r.Context(), assets.RemoteDDL)
			if err != nil {
				slogger.Error("Create Table", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			slogger.InfoContext(r.Context(), "Create Table Result", "Result", result)
			json.NewEncoder(w).Encode(registrationResponse{ ServerID: registrationRequest.ServerID, GroupID: groupID.String()})
		default:
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	libSQLConfig, err := dbconfig.LibSQLConfigFromEnv()
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode([]byte(libSQLConfig.AuthKey))
	parseResult, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	key := parseResult.(ed25519.PrivateKey)

	libSQLDatabaseURL, err := libSQLConfig.DatabaseURL(key)
	if err != nil {
		panic(err)
	}

	jwtDuration := 10 * time.Minute
	jwtCache := &jwt.TokenCache{}
	go jwt.StartJwtGeneration(jwtDuration, jwtCache, key)

	db, err := sql.Open("libsql", libSQLDatabaseURL.String());
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbCtx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()
	if err := db.PingContext(dbCtx); err != nil {
		panic(err)
	}

	reportEndpointHandler := createReportHandler(db)
	databaseEndpointHandler := createRegistrationHandler(jwtCache, db, libSQLConfig, key)

	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/databases":
			databaseEndpointHandler(w, r)
		case "/reports":
			reportEndpointHandler(w, r)
		default:
			http.Error(w, "404 Not Found", http.StatusNotFound)
		}
	}

	handler = middleware.RecordResponse(handler)
	handler = middleware.Recovery(handler)
	handler = middleware.PermissiveCORSHandler(handler)
	handler = middleware.Log(handler)
	handler = middleware.Trace(handler)

	server := http.Server {
		Addr: fmt.Sprintf(":%d", *port),
		Handler: handler,
		ReadTimeout: 1 * time.Second,
		WriteTimeout: 1 * time.Second,
		ReadHeaderTimeout: 200 * time.Millisecond,
	}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error(err.Error())
	}
}
