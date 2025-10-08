package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// TokenService manages authentication tokens
type TokenService struct {
	sessions map[string]SessionData
	mutex    sync.RWMutex
}

// SessionData represents session information
type SessionData struct {
	UserID    int
	Username  string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewTokenService creates a new token service
func NewTokenService() *TokenService {
	service := &TokenService{
		sessions: make(map[string]SessionData),
	}

	// Start cleanup goroutine to remove expired sessions
	go service.cleanupExpiredSessions()

	return service
}

// GenerateToken creates a new session token for a user
func (ts *TokenService) GenerateToken(userID int, username, email string) (string, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	// Store session data
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	sessionData := SessionData{
		UserID:    userID,
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry
	}

	ts.sessions[token] = sessionData

	return token, nil
}

// ValidateToken checks if a token is valid and returns user info
func (ts *TokenService) ValidateToken(token string) (*SessionData, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	sessionData, exists := ts.sessions[token]
	if !exists {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if time.Now().After(sessionData.ExpiresAt) {
		// Remove expired token
		ts.mutex.RUnlock()
		ts.mutex.Lock()
		delete(ts.sessions, token)
		ts.mutex.Unlock()
		ts.mutex.RLock()
		return nil, errors.New("token expired")
	}

	return &sessionData, nil
}

// InvalidateToken removes a token from the session store
func (ts *TokenService) InvalidateToken(token string) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	delete(ts.sessions, token)
}

// cleanupExpiredSessions runs periodically to clean up expired sessions
func (ts *TokenService) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ts.mutex.Lock()
		now := time.Now()
		for token, session := range ts.sessions {
			if now.After(session.ExpiresAt) {
				delete(ts.sessions, token)
			}
		}
		ts.mutex.Unlock()
	}
}
