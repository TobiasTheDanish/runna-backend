package database

import (
	"database/sql"
	"math"
	"time"

	"github.com/thc/runna-backend/internal/models"
)

func (db *DB) CreateGoal(req models.CreateGoalRequest) (*models.Goal, error) {
	query := `
		INSERT INTO goals (target_distance, start_date, end_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, target_distance, start_date, end_date, created_at, updated_at
	`

	now := time.Now()
	var goal models.Goal

	err := db.conn.QueryRow(
		query,
		req.TargetDistance,
		req.StartDate,
		req.EndDate,
		now,
		now,
	).Scan(
		&goal.ID,
		&goal.TargetDistance,
		&goal.StartDate,
		&goal.EndDate,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &goal, nil
}

func (db *DB) GetGoals() ([]models.GoalProgress, error) {
	query := `
		SELECT id, target_distance, start_date, end_date, created_at, updated_at
		FROM goals
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.GoalProgress
	for rows.Next() {
		var g models.Goal
		if err := rows.Scan(
			&g.ID,
			&g.TargetDistance,
			&g.StartDate,
			&g.EndDate,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Calculate progress
		progress, err := db.calculateGoalProgress(g)
		if err != nil {
			return nil, err
		}
		goals = append(goals, *progress)
	}

	return goals, nil
}

func (db *DB) GetGoal(id int) (*models.GoalProgress, error) {
	query := `
		SELECT id, target_distance, start_date, end_date, created_at, updated_at
		FROM goals
		WHERE id = ?
	`

	var goal models.Goal
	err := db.conn.QueryRow(query, id).Scan(
		&goal.ID,
		&goal.TargetDistance,
		&goal.StartDate,
		&goal.EndDate,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return db.calculateGoalProgress(goal)
}

func (db *DB) DeleteGoal(id int) error {
	query := `DELETE FROM goals WHERE id = ?`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (db *DB) calculateGoalProgress(goal models.Goal) (*models.GoalProgress, error) {
	// Get sessions within the goal period
	sessions, err := db.GetSessions(goal.StartDate, goal.EndDate)
	if err != nil {
		return nil, err
	}

	var totalDistance float64
	for _, s := range sessions {
		totalDistance += s.Distance
	}

	now := time.Now()
	// If current date is past end date, use end date for calculation
	calcDate := now
	if calcDate.After(goal.EndDate) {
		calcDate = goal.EndDate
	}
	// If current date is before start date, use start date (progress 0)
	if calcDate.Before(goal.StartDate) {
		calcDate = goal.StartDate
	}

	totalDuration := goal.EndDate.Sub(goal.StartDate).Hours()
	elapsedDuration := calcDate.Sub(goal.StartDate).Hours()

	var expectedDistance float64
	if totalDuration > 0 {
		expectedDistance = (elapsedDuration / totalDuration) * goal.TargetDistance
	}

	progressPercentage := (totalDistance / goal.TargetDistance) * 100
	if progressPercentage > 100 {
		progressPercentage = 100
	}

	status := "On Track"
	if totalDistance >= goal.TargetDistance {
		status = "Completed"
	} else if totalDistance < expectedDistance {
		status = "Behind"
	} else if totalDistance > expectedDistance*1.1 { // 10% buffer for "Ahead"
		status = "Ahead"
	}

	return &models.GoalProgress{
		Goal:               goal,
		CurrentDistance:    math.Round(totalDistance*100) / 100,
		ProgressPercentage: math.Round(progressPercentage*100) / 100,
		Status:             status,
		ExpectedDistance:   math.Round(expectedDistance*100) / 100,
		Sessions:           sessions,
	}, nil
}
