package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// User represents a user in the system
type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"` // Don't include in JSON responses
	FirstName       *string   `json:"first_name,omitempty"`
	LastName        *string   `json:"last_name,omitempty"`
	DateOfBirth     *string   `json:"date_of_birth,omitempty"`
	AvatarUrl       *string   `json:"avatar_url,omitempty"`
	AboutMe         *string   `json:"about_me,omitempty"`
	IsPublicProfile bool      `json:"is_public_profile"`
	CreatedAt       time.Time `json:"created_at"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// Validate validates the registration request
func (r *RegisterRequest) Validate() error {
	// Username validation
	if r.Username == "" {
		return errors.New("username is required")
	}

	if len(r.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	// Email validation
	if r.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	// Password validation
	if r.Password == "" {
		return errors.New("password is required")
	}

	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	// Name validation
	if strings.TrimSpace(r.FirstName) == "" {
		return errors.New("first name is required")
	}

	if strings.TrimSpace(r.LastName) == "" {
		return errors.New("last name is required")
	}

	return nil
}

// Validate validates the login request
func (l *LoginRequest) Validate() error {
	if l.Email == "" {
		return errors.New("email is required")
	}

	if l.Password == "" {
		return errors.New("password is required")
	}

	return nil
}
