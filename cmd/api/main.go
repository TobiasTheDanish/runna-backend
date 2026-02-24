package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/thc/runna-backend/internal/database"
	"github.com/thc/runna-backend/internal/handlers"
	"github.com/thc/runna-backend/internal/middleware"
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

	h := handlers.New(db)

	// Initialize Strava service
	stravaService := services.NewStravaService(db)
	h.SetStravaService(stravaService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sessions", h.CreateSession)
	mux.HandleFunc("GET /api/sessions", h.GetSessions)
	mux.HandleFunc("GET /api/sessions/{id}", h.GetSession)
	mux.HandleFunc("PUT /api/sessions/{id}", h.UpdateSession)

	// Goal routes
	mux.HandleFunc("POST /api/goals", h.CreateGoal)
	mux.HandleFunc("GET /api/goals", h.GetGoals)
	mux.HandleFunc("GET /api/goals/{id}", h.GetGoal)
	mux.HandleFunc("DELETE /api/goals/{id}", h.DeleteGoal)

	// Strava webhook routes
	mux.HandleFunc("GET /api/webhooks/strava", h.VerifyWebhook)
	mux.HandleFunc("POST /api/webhooks/strava", h.ReceiveWebhook)

	// Strava OAuth routes
	mux.HandleFunc("POST /api/strava/connect", h.ConnectStrava)
	mux.HandleFunc("GET /api/strava/status", h.GetStravaStatus)
	mux.HandleFunc("DELETE /api/strava/disconnect", h.DisconnectStrava)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Apply middleware: logging first, then CORS
	handler := middleware.Logging(enableCORS(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
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
