package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/thc/runna-backend/internal/database"
	"github.com/thc/runna-backend/internal/models"
)

type Handler struct {
	db            *database.DB
	stravaService interface {
		ProcessWebhookEvent(event models.WebhookEvent) error
	}
}

func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) SetStravaService(service interface {
	ProcessWebhookEvent(event models.WebhookEvent) error
}) {
	h.stravaService = service
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] CreateSession: Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Distance <= 0 {
		log.Printf("[WARN] CreateSession: Invalid distance: %f", req.Distance)
		http.Error(w, "Distance must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.Duration <= 0 {
		log.Printf("[WARN] CreateSession: Invalid duration: %d", req.Duration)
		http.Error(w, "Duration must be greater than 0", http.StatusBadRequest)
		return
	}

	session, err := h.db.CreateSession(req)
	if err != nil {
		log.Printf("[ERROR] CreateSession: Database error: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] CreateSession: Created session id=%d", session.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		startDate = time.Now().AddDate(0, -1, 0)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			log.Printf("[WARN] GetSessions: Invalid start_date format: %s, error: %v", startDateStr, err)
			http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	if endDateStr == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			log.Printf("[WARN] GetSessions: Invalid end_date format: %s, error: %v", endDateStr, err)
			http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	sessions, err := h.db.GetSessions(startDate, endDate)
	if err != nil {
		log.Printf("[ERROR] GetSessions: Database error: %v", err)
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	if sessions == nil {
		sessions = []models.Session{}
	}

	log.Printf("[INFO] GetSessions: Retrieved %d sessions for period %s to %s", len(sessions), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[WARN] GetSession: Invalid session ID format: %s, error: %v", idStr, err)
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := h.db.GetSession(id)
	if err != nil {
		log.Printf("[ERROR] GetSession: Database error for id=%d: %v", id, err)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	log.Printf("[INFO] GetSession: Retrieved session id=%d", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[WARN] UpdateSession: Invalid session ID format: %s, error: %v", idStr, err)
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] UpdateSession: Failed to decode request body for id=%d: %v", id, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Distance <= 0 {
		log.Printf("[WARN] UpdateSession: Invalid distance for id=%d: %f", id, req.Distance)
		http.Error(w, "Distance must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.Duration <= 0 {
		log.Printf("[WARN] UpdateSession: Invalid duration for id=%d: %d", id, req.Duration)
		http.Error(w, "Duration must be greater than 0", http.StatusBadRequest)
		return
	}

	session, err := h.db.UpdateSession(id, req)
	if err != nil {
		log.Printf("[ERROR] UpdateSession: Database error for id=%d: %v", id, err)
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] UpdateSession: Updated session id=%d", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
