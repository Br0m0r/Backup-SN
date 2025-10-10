package models

import "time"

// User represents a user profile
type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	FirstName       *string   `json:"first_name,omitempty"`
	LastName        *string   `json:"last_name,omitempty"`
	DateOfBirth     *string   `json:"date_of_birth,omitempty"`
	AvatarPath      *string   `json:"avatar_path,omitempty"`
	Nickname        *string   `json:"nickname,omitempty"`
	AboutMe         *string   `json:"about_me,omitempty"`
	IsPublicProfile bool      `json:"is_public_profile"`
	CreatedAt       time.Time `json:"created_at"`
}

// UpdateProfileRequest represents profile update payload
type UpdateProfileRequest struct {
	FirstName       *string `json:"first_name,omitempty"`
	LastName        *string `json:"last_name,omitempty"`
	DateOfBirth     *string `json:"date_of_birth,omitempty"`
	Nickname        *string `json:"nickname,omitempty"`
	AboutMe         *string `json:"about_me,omitempty"`
	IsPublicProfile *bool   `json:"is_public_profile,omitempty"`
}

// Follow represents a follow relationship
type Follow struct {
	ID          int       `json:"id"`
	FollowerID  int       `json:"follower_id"`
	FollowingID int       `json:"following_id"`
	Status      string    `json:"status"` // "pending" or "accepted"
	CreatedAt   time.Time `json:"created_at"`
}

// FollowRequest represents a follow action request
type FollowRequest struct {
	UserID int `json:"user_id"`
}
