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
	hub    *Hub
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

	hub := newHub()
	go hub.run()

	apiCfg := apiConfig{
		db:     dbQueries,
		secret: secret,
		hub:    hub,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLoginUser)

	mux.HandleFunc("POST /api/chats", apiCfg.middlewareAuth(apiCfg.handlerChat))
	mux.HandleFunc("GET /api/chats/ws", apiCfg.handlerChatWS)
	mux.HandleFunc("POST /api/chats/{chat_id}/read", apiCfg.middlewareAuth(apiCfg.handlerMarkAsRead))

	mux.HandleFunc("GET /api/notifications", apiCfg.middlewareAuth(apiCfg.handlerNotifications))

	mux.HandleFunc("POST /api/friends", apiCfg.middlewareAuth(apiCfg.handlerFriends))
	mux.HandleFunc("GET /api/friends", apiCfg.middlewareAuth(apiCfg.handlerGetFriends))
	mux.HandleFunc("DELETE /api/friends", apiCfg.middlewareAuth(apiCfg.handlerDeleteFriend))

	mux.HandleFunc("POST /api/online", apiCfg.middlewareAuth(apiCfg.handlerSetOnline))
	mux.HandleFunc("POST /api/offline", apiCfg.middlewareAuth(apiCfg.handlerSetOffline))

	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
