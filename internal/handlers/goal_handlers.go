package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/thc/runna-backend/internal/models"
)

func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req models.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetDistance <= 0 {
		http.Error(w, "Target distance must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.EndDate.Before(req.StartDate) {
		http.Error(w, "End date must be after start date", http.StatusBadRequest)
		return
	}

	goal, err := h.db.CreateGoal(req)
	if err != nil {
		http.Error(w, "Failed to create goal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) GetGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := h.db.GetGoals()
	if err != nil {
		http.Error(w, "Failed to get goals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func (h *Handler) GetGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	goal, err := h.db.GetGoal(id)
	if err != nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteGoal(id); err != nil {
		http.Error(w, "Failed to delete goal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
