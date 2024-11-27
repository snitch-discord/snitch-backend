package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"
	"snitch/snitchbe/pkg/ctxutil"
	"snitch/snitchbe/pkg/middleware"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Report struct {
	Text string `json:"reportText"`
	ReporterId int `json:"reporterId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	ReportedUserId int `json:"reporteduserId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
	ServerId int `json:"serverId,string"` // we need to tell go that our number is encoded as a string, hence ',string'
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


// func createDatabase(newDatabaseName string, sqlConfig dbconfig.LibSQLConfig) string {
// 	ctx := context.TODO();
// 	requestUrl := sqlConfig.DatabaseUrl + "/v1/organizations/"

// }

func createDatabaseHandler(tokenCache *jwt.TokenCache, libsqlUrl string) http.HandlerFunc {
	httpClient := &http.Client{}

	return func(w http.ResponseWriter, r *http.Request) {
		slogger, ok := ctxutil.Value[*slog.Logger](r.Context())
		if !ok {
			slogger = slog.Default()
		}

		switch (r.Method) {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			
			request, _ := http.NewRequestWithContext(r.Context(), "GET", libsqlUrl + "/v1", nil)
			request.Header.Add("Authorization", "Bearer " + tokenCache.Get())
			response, err := httpClient.Do(request)

			if err != nil {
				slogger.Error("Error", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer response.Body.Close()
			_, err = io.Copy(w, response.Body)
			if err != nil {
				slogger.Error("Error", "Error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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
	ctx := context.Background()

	block, _ := pem.Decode([]byte(libSQLConfig.AuthKey))
	parseResult, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	key := parseResult.(ed25519.PrivateKey)

	jwtDuration := 10 * time.Minute
	jwtCache := &jwt.TokenCache{}
	go jwt.StartJwtGeneration(jwtDuration, jwtCache, key)

	db, err:= sql.Open("libsql", libSQLConfig.DatabaseURL(key));
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbCtx, cancel := context.WithTimeout(ctx, time.Second * 5)
	defer cancel()
	if err := db.PingContext(dbCtx); err != nil {
		panic(err)
	}

	reportEndpointHandler := createReportHandler(db)
	databaseEndpointHandler := createDatabaseHandler(jwtCache, libSQLConfig.HttpURL())

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
