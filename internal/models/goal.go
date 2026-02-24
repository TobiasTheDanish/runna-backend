package models

import "time"

type Goal struct {
	ID             int64     `json:"id"`
	TargetDistance float64   `json:"target_distance"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateGoalRequest struct {
	TargetDistance float64   `json:"target_distance"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}

type GoalProgress struct {
	Goal
	CurrentDistance    float64   `json:"current_distance"`
	ProgressPercentage float64   `json:"progress_percentage"`
	Status             string    `json:"status"` // "On Track", "Behind", "Ahead", "Completed"
	ExpectedDistance   float64   `json:"expected_distance"`
	Sessions           []Session `json:"sessions,omitempty"`
}
