package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/thc/runna-backend/internal/database"
	"github.com/thc/runna-backend/internal/handlers"
	"github.com/thc/runna-backend/internal/services"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	handler := handlers.New(db)

	// Initialize Strava service
	stravaService := services.NewStravaService(db)
	handler.SetStravaService(stravaService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sessions", handler.CreateSession)
	mux.HandleFunc("GET /api/sessions", handler.GetSessions)
	mux.HandleFunc("GET /api/sessions/{id}", handler.GetSession)
	mux.HandleFunc("PUT /api/sessions/{id}", handler.UpdateSession)

	// Goal routes
	mux.HandleFunc("POST /api/goals", handler.CreateGoal)
	mux.HandleFunc("GET /api/goals", handler.GetGoals)
	mux.HandleFunc("GET /api/goals/{id}", handler.GetGoal)
	mux.HandleFunc("DELETE /api/goals/{id}", handler.DeleteGoal)

	// Strava webhook routes
	mux.HandleFunc("GET /api/webhooks/strava", handler.VerifyWebhook)
	mux.HandleFunc("POST /api/webhooks/strava", handler.ReceiveWebhook)

	// Strava OAuth routes
	mux.HandleFunc("POST /api/strava/connect", handler.ConnectStrava)
	mux.HandleFunc("GET /api/strava/status", handler.GetStravaStatus)
	mux.HandleFunc("DELETE /api/strava/disconnect", handler.DisconnectStrava)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	corsHandler := enableCORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
