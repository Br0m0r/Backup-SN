package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// TokenService manages authentication tokens and sessions
type TokenService struct {
	database *sql.DB
}

// SessionData represents session information stored in database
type SessionData struct {
	ID        int
	UserID    int
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewTokenService creates a new token service
func NewTokenService(db *sql.DB) *TokenService {
	service := &TokenService{
		database: db,
	}

	// Start cleanup goroutine to remove expired sessions
	go service.cleanupExpiredSessions()

	return service
}

// GenerateToken creates a new session token for a user and stores it in the database
func (ts *TokenService) GenerateToken(userID int, username, email string) (string, error) {
	// Generate random token (64 character hex string)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	// Calculate expiration (24 hours from now)
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	// Store session in database
	query := `INSERT INTO sessions (user_id, token, created_at, expires_at) VALUES (?, ?, ?, ?)`
	_, err := ts.database.Exec(query, userID, token, now, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken checks if a token is valid in the database and returns user info
func (ts *TokenService) ValidateToken(token string) (*SessionData, error) {
	query := `
		SELECT id, user_id, token, created_at, expires_at 
		FROM sessions 
		WHERE token = ? AND expires_at > datetime('now')
	`

	var session SessionData
	err := ts.database.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid or expired token")
		}
		return nil, err
	}

	return &session, nil
}

// InvalidateToken removes a token from the database (logout)
func (ts *TokenService) InvalidateToken(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := ts.database.Exec(query, token)
	return err
}

// cleanupExpiredSessions runs periodically to clean up expired sessions from database
func (ts *TokenService) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		query := `DELETE FROM sessions WHERE expires_at < datetime('now')`
		ts.database.Exec(query)
	}
}
