package models

import (
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
	AvatarPath      *string   `json:"avatar_path,omitempty"`
	Nickname        *string   `json:"nickname,omitempty"`
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
