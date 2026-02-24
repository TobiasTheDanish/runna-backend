package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/thc/runna-backend/internal/models"
)

// VerifyWebhook handles Strava webhook subscription verification (GET request)
func (h *Handler) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	// Get verify token from environment
	verifyToken := os.Getenv("STRAVA_VERIFY_TOKEN")
	if verifyToken == "" {
		log.Println("STRAVA_VERIFY_TOKEN not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	// Verify token and mode
	if mode == "subscribe" && token == verifyToken {
		log.Println("Webhook verified successfully")

		// Respond with challenge
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"hub.challenge": challenge,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Invalid token or mode
	log.Printf("Webhook verification failed: mode=%s, token=%s", mode, token)
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// ReceiveWebhook handles incoming webhook events from Strava (POST request)
func (h *Handler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	var event models.WebhookEvent

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("Failed to decode webhook event: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received webhook event: type=%s, aspect=%s, object_id=%d, owner_id=%d",
		event.ObjectType, event.AspectType, event.ObjectID, event.OwnerID)

	// Queue the event for async processing
	// For now, we'll use a simple goroutine. In production, use a proper job queue.
	go h.processWebhookEvent(event)

	// Respond immediately with 200 OK (must respond within 2 seconds)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

// processWebhookEvent processes webhook events asynchronously
func (h *Handler) processWebhookEvent(event models.WebhookEvent) {
	if h.stravaService == nil {
		log.Println("Strava service not initialized")
		return
	}

	// Process the event using the Strava service
	if err := h.stravaService.ProcessWebhookEvent(event); err != nil {
		log.Printf("Failed to process webhook event: %v", err)
	}
}
