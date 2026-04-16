package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/denisnosik/cli-messenger/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db     *database.Queries
	secret string
}

func Run() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	secret := os.Getenv("SECRET")

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
		db:     dbQueries,
		secret: secret,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
