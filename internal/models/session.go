package models

import "time"

type Session struct {
	ID        int64     `json:"id"`
	Date      time.Time `json:"date"`
	Distance  float64   `json:"distance"`
	Duration  int       `json:"duration"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSessionRequest struct {
	Date     time.Time `json:"date"`
	Distance float64   `json:"distance"`
	Duration int       `json:"duration"`
	Notes    string    `json:"notes"`
}
