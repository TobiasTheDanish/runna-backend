package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/thc/runna-backend/internal/models"
	"github.com/thc/runna-backend/internal/services"
)

// ConnectStrava handles OAuth token exchange
func (h *Handler) ConnectStrava(w http.ResponseWriter, r *http.Request) {
	var req models.StravaConnectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, "Authorization code is required", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	client := services.NewStravaClient()
	tokenResp, err := client.ExchangeToken(req.Code)
	if err != nil {
		log.Printf("Failed to exchange token: %v", err)
		http.Error(w, "Failed to connect to Strava", http.StatusInternalServerError)
		return
	}

	// Store connection in database
	conn := models.StravaConnection{
		StravaAthleteID: tokenResp.Athlete.ID,
		AccessToken:     tokenResp.AccessToken,
		RefreshToken:    tokenResp.RefreshToken,
		TokenExpiresAt:  time.Unix(tokenResp.ExpiresAt, 0),
	}

	createdConn, err := h.db.CreateStravaConnection(conn)
	if err != nil {
		log.Printf("Failed to store connection: %v", err)
		http.Error(w, "Failed to store connection", http.StatusInternalServerError)
		return
	}

	log.Printf("Strava connection created for athlete %d", createdConn.StravaAthleteID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":           true,
		"strava_athlete_id": createdConn.StravaAthleteID,
		"connected_at":      createdConn.ConnectedAt,
	})
}

// GetStravaStatus returns the current Strava connection status
func (h *Handler) GetStravaStatus(w http.ResponseWriter, r *http.Request) {
	// For MVP, we just check if any connection exists (single-user)
	conn, err := h.db.GetStravaConnection()
	if err != nil {
		log.Printf("Failed to get connection: %v", err)
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	status := models.StravaConnectionStatus{
		Connected: conn != nil,
	}

	if conn != nil {
		status.StravaAthleteID = conn.StravaAthleteID
		status.ConnectedAt = conn.ConnectedAt
		status.LastSync = conn.LastSync
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// DisconnectStrava removes the Strava connection
func (h *Handler) DisconnectStrava(w http.ResponseWriter, r *http.Request) {
	// For MVP, disconnect the first/only connection
	conn, err := h.db.GetStravaConnection()
	if err != nil {
		log.Printf("Failed to get connection: %v", err)
		http.Error(w, "Failed to disconnect", http.StatusInternalServerError)
		return
	}

	if conn == nil {
		http.Error(w, "No Strava connection found", http.StatusNotFound)
		return
	}

	err = h.db.DeleteStravaConnection(conn.StravaAthleteID)
	if err != nil {
		log.Printf("Failed to delete connection: %v", err)
		http.Error(w, "Failed to disconnect", http.StatusInternalServerError)
		return
	}

	log.Printf("Strava connection deleted for athlete %d", conn.StravaAthleteID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Disconnected from Strava",
	})
}
