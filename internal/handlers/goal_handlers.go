package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/thc/runna-backend/internal/models"
)

func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req models.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] CreateGoal: Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetDistance <= 0 {
		log.Printf("[WARN] CreateGoal: Invalid target distance: %f", req.TargetDistance)
		http.Error(w, "Target distance must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.EndDate.Before(req.StartDate) {
		log.Printf("[WARN] CreateGoal: End date before start date: start=%s, end=%s", req.StartDate, req.EndDate)
		http.Error(w, "End date must be after start date", http.StatusBadRequest)
		return
	}

	goal, err := h.db.CreateGoal(req)
	if err != nil {
		log.Printf("[ERROR] CreateGoal: Database error: %v", err)
		http.Error(w, "Failed to create goal", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] CreateGoal: Created goal id=%d, target=%.2fkm", goal.ID, goal.TargetDistance)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) GetGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := h.db.GetGoals()
	if err != nil {
		log.Printf("[ERROR] GetGoals: Database error: %v", err)
		http.Error(w, "Failed to get goals", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] GetGoals: Retrieved %d goals", len(goals))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func (h *Handler) GetGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[WARN] GetGoal: Invalid goal ID format: %s, error: %v", idStr, err)
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	goal, err := h.db.GetGoal(id)
	if err != nil {
		log.Printf("[ERROR] GetGoal: Database error for id=%d: %v", id, err)
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	log.Printf("[INFO] GetGoal: Retrieved goal id=%d", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[WARN] DeleteGoal: Invalid goal ID format: %s, error: %v", idStr, err)
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteGoal(id); err != nil {
		log.Printf("[ERROR] DeleteGoal: Database error for id=%d: %v", id, err)
		http.Error(w, "Failed to delete goal", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] DeleteGoal: Deleted goal id=%d", id)
	w.WriteHeader(http.StatusNoContent)
}
