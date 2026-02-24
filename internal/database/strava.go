package database

import (
	"database/sql"
	"time"

	"github.com/thc/runna-backend/internal/models"
)

// CreateStravaConnection stores a new Strava connection
func (db *DB) CreateStravaConnection(conn models.StravaConnection) (*models.StravaConnection, error) {
	query := `
		INSERT INTO strava_connections (user_id, strava_athlete_id, access_token, refresh_token, token_expires_at, connected_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, user_id, strava_athlete_id, access_token, refresh_token, token_expires_at, connected_at, last_sync
	`

	var result models.StravaConnection
	err := db.conn.QueryRow(
		query,
		conn.UserID,
		conn.StravaAthleteID,
		conn.AccessToken,
		conn.RefreshToken,
		conn.TokenExpiresAt,
		time.Now(),
	).Scan(
		&result.ID,
		&result.UserID,
		&result.StravaAthleteID,
		&result.AccessToken,
		&result.RefreshToken,
		&result.TokenExpiresAt,
		&result.ConnectedAt,
		&result.LastSync,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStravaConnectionByAthleteID retrieves a Strava connection by athlete ID
func (db *DB) GetStravaConnectionByAthleteID(athleteID int64) (*models.StravaConnection, error) {
	query := `
		SELECT id, user_id, strava_athlete_id, access_token, refresh_token, token_expires_at, connected_at, last_sync
		FROM strava_connections
		WHERE strava_athlete_id = ?
	`

	var conn models.StravaConnection
	err := db.conn.QueryRow(query, athleteID).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.StravaAthleteID,
		&conn.AccessToken,
		&conn.RefreshToken,
		&conn.TokenExpiresAt,
		&conn.ConnectedAt,
		&conn.LastSync,
	)

	if err != nil {
		return nil, err
	}

	return &conn, nil
}

// GetStravaConnection retrieves the first Strava connection (for single-user MVP)
func (db *DB) GetStravaConnection() (*models.StravaConnection, error) {
	query := `
		SELECT id, user_id, strava_athlete_id, access_token, refresh_token, token_expires_at, connected_at, last_sync
		FROM strava_connections
		LIMIT 1
	`

	var conn models.StravaConnection
	err := db.conn.QueryRow(query).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.StravaAthleteID,
		&conn.AccessToken,
		&conn.RefreshToken,
		&conn.TokenExpiresAt,
		&conn.ConnectedAt,
		&conn.LastSync,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conn, nil
}

// UpdateStravaTokens updates access and refresh tokens
func (db *DB) UpdateStravaTokens(athleteID int64, accessToken, refreshToken string, expiresAt time.Time) error {
	query := `
		UPDATE strava_connections
		SET access_token = ?, refresh_token = ?, token_expires_at = ?
		WHERE strava_athlete_id = ?
	`

	_, err := db.conn.Exec(query, accessToken, refreshToken, expiresAt, athleteID)
	return err
}

// DeleteStravaConnection removes a Strava connection
func (db *DB) DeleteStravaConnection(athleteID int64) error {
	query := `DELETE FROM strava_connections WHERE strava_athlete_id = ?`
	_, err := db.conn.Exec(query, athleteID)
	return err
}

// CreateStravaSession creates a session from Strava activity
func (db *DB) CreateStravaSession(session models.Session) (*models.Session, error) {
	query := `
		INSERT INTO sessions (date, distance, duration, notes, strava_activity_id, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'strava', ?, ?)
		RETURNING id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
	`

	now := time.Now()
	var result models.Session

	err := db.conn.QueryRow(
		query,
		session.Date,
		session.Distance,
		session.Duration,
		session.Notes,
		session.StravaActivityID,
		now,
		now,
	).Scan(
		&result.ID,
		&result.Date,
		&result.Distance,
		&result.Duration,
		&result.Notes,
		&result.StravaActivityID,
		&result.Source,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSessionByStravaActivityID retrieves a session by Strava activity ID
func (db *DB) GetSessionByStravaActivityID(activityID int64) (*models.Session, error) {
	query := `
		SELECT id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
		FROM sessions
		WHERE strava_activity_id = ?
	`

	var session models.Session
	err := db.conn.QueryRow(query, activityID).Scan(
		&session.ID,
		&session.Date,
		&session.Distance,
		&session.Duration,
		&session.Notes,
		&session.StravaActivityID,
		&session.Source,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateStravaSession updates a session from Strava activity
func (db *DB) UpdateStravaSession(activityID int64, session models.Session) (*models.Session, error) {
	query := `
		UPDATE sessions
		SET date = ?, distance = ?, duration = ?, notes = ?, updated_at = ?
		WHERE strava_activity_id = ?
		RETURNING id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
	`

	now := time.Now()
	var result models.Session

	err := db.conn.QueryRow(
		query,
		session.Date,
		session.Distance,
		session.Duration,
		session.Notes,
		now,
		activityID,
	).Scan(
		&result.ID,
		&result.Date,
		&result.Distance,
		&result.Duration,
		&result.Notes,
		&result.StravaActivityID,
		&result.Source,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteSessionByStravaActivityID deletes a session by Strava activity ID
func (db *DB) DeleteSessionByStravaActivityID(activityID int64) error {
	query := `DELETE FROM sessions WHERE strava_activity_id = ?`
	_, err := db.conn.Exec(query, activityID)
	return err
}
