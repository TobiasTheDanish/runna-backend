package models

import "time"

// StravaConnection represents a connection to a Strava account
type StravaConnection struct {
	ID              int64      `json:"id"`
	UserID          *int64     `json:"user_id,omitempty"` // Future-proofing for multi-user
	StravaAthleteID int64      `json:"strava_athlete_id"`
	AccessToken     string     `json:"-"` // Never expose in JSON
	RefreshToken    string     `json:"-"` // Never expose in JSON
	TokenExpiresAt  time.Time  `json:"token_expires_at"`
	ConnectedAt     time.Time  `json:"connected_at"`
	LastSync        *time.Time `json:"last_sync,omitempty"`
}

// StravaConnectionStatus is returned to the frontend
type StravaConnectionStatus struct {
	Connected       bool       `json:"connected"`
	StravaAthleteID int64      `json:"strava_athlete_id,omitempty"`
	ConnectedAt     time.Time  `json:"connected_at,omitempty"`
	LastSync        *time.Time `json:"last_sync,omitempty"`
}

// WebhookEvent represents a Strava webhook event
type WebhookEvent struct {
	AspectType     string                 `json:"aspect_type"`     // "create", "update", or "delete"
	EventTime      int64                  `json:"event_time"`      // Unix timestamp
	ObjectID       int64                  `json:"object_id"`       // Activity or athlete ID
	ObjectType     string                 `json:"object_type"`     // "activity" or "athlete"
	OwnerID        int64                  `json:"owner_id"`        // Athlete ID
	SubscriptionID int                    `json:"subscription_id"` // Webhook subscription ID
	Updates        map[string]interface{} `json:"updates"`         // Changed fields
}

// StravaActivity represents a Strava activity from the API
type StravaActivity struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`        // "Run", "Ride", etc.
	Distance   float64   `json:"distance"`    // meters
	MovingTime int       `json:"moving_time"` // seconds
	StartDate  time.Time `json:"start_date"`
	Private    bool      `json:"private"`
}

// StravaTokenResponse represents the OAuth token response
type StravaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Athlete      struct {
		ID int64 `json:"id"`
	} `json:"athlete"`
}

// StravaConnectRequest represents the OAuth authorization code
type StravaConnectRequest struct {
	Code string `json:"code"`
}
