package services

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/thc/runna-backend/internal/crypto"
	"github.com/thc/runna-backend/internal/database"
	"github.com/thc/runna-backend/internal/models"
)

type StravaService struct {
	db     *database.DB
	client *StravaClient
}

func NewStravaService(db *database.DB) *StravaService {
	return &StravaService{
		db:     db,
		client: NewStravaClient(),
	}
}

// ProcessWebhookEvent processes incoming Strava webhook events
func (s *StravaService) ProcessWebhookEvent(event models.WebhookEvent) error {
	switch event.ObjectType {
	case "activity":
		return s.processActivityEvent(event)
	case "athlete":
		return s.processAthleteEvent(event)
	default:
		log.Printf("Unknown object type: %s", event.ObjectType)
		return nil
	}
}

// processActivityEvent handles activity create/update/delete events
func (s *StravaService) processActivityEvent(event models.WebhookEvent) error {
	switch event.AspectType {
	case "create":
		return s.ProcessActivityCreated(event.ObjectID, event.OwnerID)
	case "update":
		return s.ProcessActivityUpdated(event.ObjectID, event.OwnerID, event.Updates)
	case "delete":
		return s.ProcessActivityDeleted(event.ObjectID)
	default:
		log.Printf("Unknown aspect type: %s", event.AspectType)
		return nil
	}
}

// processAthleteEvent handles athlete deauthorization events
func (s *StravaService) processAthleteEvent(event models.WebhookEvent) error {
	if event.AspectType == "update" {
		if authorized, ok := event.Updates["authorized"].(string); ok && authorized == "false" {
			return s.ProcessAthleteDeauthorized(event.OwnerID)
		}
	}
	return nil
}

// ProcessActivityCreated handles new activity creation
func (s *StravaService) ProcessActivityCreated(activityID, ownerID int64) error {
	log.Printf("Processing activity created: activityID=%d, ownerID=%d", activityID, ownerID)

	// Get athlete's connection
	conn, err := s.db.GetStravaConnectionByAthleteID(ownerID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	if conn == nil {
		log.Printf("No connection found for athlete %d", ownerID)
		return nil
	}

	// Check if token needs refresh
	accessToken, err := s.ensureValidToken(conn)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Fetch activity details
	activity, err := s.client.GetActivity(accessToken, activityID)
	if err != nil {
		return fmt.Errorf("failed to fetch activity: %w", err)
	}

	// Only process running activities
	if activity.Type != "Run" {
		log.Printf("Skipping non-running activity: type=%s", activity.Type)
		return nil
	}

	// Check if activity already exists
	existing, err := s.db.GetSessionByStravaActivityID(activityID)
	if err != nil {
		return fmt.Errorf("failed to check existing session: %w", err)
	}
	if existing != nil {
		log.Printf("Session already exists for activity %d", activityID)
		return nil
	}

	// Convert Strava activity to session
	session := models.Session{
		Date:             activity.StartDate,
		Distance:         activity.Distance / 1000, // Convert meters to km
		Duration:         activity.MovingTime,
		Notes:            activity.Name,
		StravaActivityID: &activityID,
		Source:           "strava",
	}

	// Create session
	createdSession, err := s.db.CreateStravaSession(session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("Created session %d from Strava activity %d", createdSession.ID, activityID)
	return nil
}

// ProcessActivityUpdated handles activity updates
func (s *StravaService) ProcessActivityUpdated(activityID, ownerID int64, updates map[string]interface{}) error {
	log.Printf("Processing activity updated: activityID=%d, ownerID=%d, updates=%v", activityID, ownerID, updates)

	// Get athlete's connection
	conn, err := s.db.GetStravaConnectionByAthleteID(ownerID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	if conn == nil {
		log.Printf("No connection found for athlete %d", ownerID)
		return nil
	}

	// Check if token needs refresh
	accessToken, err := s.ensureValidToken(conn)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Fetch updated activity details
	activity, err := s.client.GetActivity(accessToken, activityID)
	if err != nil {
		return fmt.Errorf("failed to fetch activity: %w", err)
	}

	// Check if activity type changed to non-running
	if activity.Type != "Run" {
		log.Printf("Activity type changed to non-running, deleting session: type=%s", activity.Type)
		return s.ProcessActivityDeleted(activityID)
	}

	// Handle privacy changes (activity becomes private for apps without activity:read_all)
	if private, ok := updates["private"].(string); ok && private == "true" {
		log.Printf("Activity became private, treating as delete")
		return s.ProcessActivityDeleted(activityID)
	}

	// Get existing session
	session, err := s.db.GetSessionByStravaActivityID(activityID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		// Session doesn't exist, treat as create
		log.Printf("Session doesn't exist, treating update as create")
		return s.ProcessActivityCreated(activityID, ownerID)
	}

	// Update session with new data
	updatedSession := models.Session{
		Date:     activity.StartDate,
		Distance: activity.Distance / 1000, // Convert meters to km
		Duration: activity.MovingTime,
		Notes:    activity.Name,
	}

	_, err = s.db.UpdateStravaSession(activityID, updatedSession)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	log.Printf("Updated session for activity %d", activityID)
	return nil
}

// ProcessActivityDeleted handles activity deletion
func (s *StravaService) ProcessActivityDeleted(activityID int64) error {
	log.Printf("Processing activity deleted: activityID=%d", activityID)

	err := s.db.DeleteSessionByStravaActivityID(activityID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	log.Printf("Deleted session for activity %d", activityID)
	return nil
}

// ProcessAthleteDeauthorized handles athlete deauthorization
func (s *StravaService) ProcessAthleteDeauthorized(athleteID int64) error {
	log.Printf("Processing athlete deauthorized: athleteID=%d", athleteID)

	err := s.db.DeleteStravaConnection(athleteID)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	log.Printf("Deleted connection for athlete %d", athleteID)
	return nil
}

// ensureValidToken checks if token is expired and refreshes if necessary
// Returns the decrypted access token ready for use
func (s *StravaService) ensureValidToken(conn *models.StravaConnection) (string, error) {
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return "", fmt.Errorf("ENCRYPTION_KEY not set")
	}

	// Decrypt the access token
	accessToken, err := crypto.Decrypt(conn.AccessToken, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	// Check if token is expired or expires soon (within 5 minutes)
	if time.Now().Add(5 * time.Minute).Before(conn.TokenExpiresAt) {
		// Token is still valid, return decrypted token
		return accessToken, nil
	}

	log.Printf("Token expired or expiring soon, refreshing for athlete %d", conn.StravaAthleteID)

	// Decrypt refresh token for API call
	refreshToken, err := crypto.Decrypt(conn.RefreshToken, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Refresh token with Strava API
	tokenResp, err := s.client.RefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	// Encrypt new tokens before storing
	encryptedAccessToken, err := crypto.Encrypt(tokenResp.AccessToken, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt new access token: %w", err)
	}

	encryptedRefreshToken, err := crypto.Encrypt(tokenResp.RefreshToken, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt new refresh token: %w", err)
	}

	// Update encrypted tokens in database
	expiresAt := time.Unix(tokenResp.ExpiresAt, 0)
	err = s.db.UpdateStravaTokens(
		conn.StravaAthleteID,
		encryptedAccessToken,
		encryptedRefreshToken,
		expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to update tokens: %w", err)
	}

	log.Printf("Token refreshed successfully for athlete %d", conn.StravaAthleteID)
	// Return the new decrypted access token
	return tokenResp.AccessToken, nil
}
