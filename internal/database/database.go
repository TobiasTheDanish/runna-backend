package database

import (
	"database/sql"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/thc/runna-backend/internal/models"
)

type DB struct {
	conn *sql.DB
}

func New(dbURL string) (*DB, error) {
	conn, err := sql.Open("libsql", dbURL)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Init() error {
	query := `
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATETIME NOT NULL,
			distance REAL NOT NULL,
			duration INTEGER NOT NULL,
			notes TEXT,
			strava_activity_id INTEGER UNIQUE,
			source TEXT DEFAULT 'manual',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS goals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			target_distance REAL NOT NULL,
			start_date DATETIME NOT NULL,
			end_date DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS strava_connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			strava_athlete_id INTEGER NOT NULL UNIQUE,
			access_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			token_expires_at DATETIME NOT NULL,
			connected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_sync DATETIME
		);
	`
	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) CreateSession(req models.CreateSessionRequest) (*models.Session, error) {
	query := `
		INSERT INTO sessions (date, distance, duration, notes, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'manual', ?, ?)
		RETURNING id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
	`

	now := time.Now()
	var session models.Session

	err := db.conn.QueryRow(
		query,
		req.Date,
		req.Distance,
		req.Duration,
		req.Notes,
		now,
		now,
	).Scan(
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

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DB) GetSessions(startDate, endDate time.Time) ([]models.Session, error) {
	query := `
		SELECT id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
		FROM sessions
		WHERE date >= ? AND date <= ?
		ORDER BY date DESC
	`

	rows, err := db.conn.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
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
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (db *DB) GetSession(id int) (*models.Session, error) {
	query := `
		SELECT id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
		FROM sessions
		WHERE id = ?
	`

	var session models.Session
	err := db.conn.QueryRow(query, id).Scan(
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

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DB) UpdateSession(id int, req models.CreateSessionRequest) (*models.Session, error) {
	query := `
		UPDATE sessions
		SET date = ?, distance = ?, duration = ?, notes = ?, updated_at = ?
		WHERE id = ?
		RETURNING id, date, distance, duration, notes, strava_activity_id, source, created_at, updated_at
	`

	now := time.Now()
	var session models.Session

	err := db.conn.QueryRow(
		query,
		req.Date,
		req.Distance,
		req.Duration,
		req.Notes,
		now,
		id,
	).Scan(
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

	if err != nil {
		return nil, err
	}

	return &session, nil
}
