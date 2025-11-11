package models

import "time"

// User represents a user profile (only fields owned by User Service)
type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	AvatarPath      *string   `json:"avatar_path,omitempty"`
	Nickname        *string   `json:"nickname,omitempty"`
	AboutMe         *string   `json:"about_me,omitempty"`
	IsPublicProfile bool      `json:"is_public_profile"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PublicProfile returns a user profile without sensitive information
func (u *User) PublicProfile() *User {
	return &User{
		ID:              u.ID,
		Username:        u.Username,
		AvatarPath:      u.AvatarPath,
		Nickname:        u.Nickname,
		AboutMe:         u.AboutMe,
		IsPublicProfile: u.IsPublicProfile,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UpdateProfileRequest represents profile update payload (only fields User Service owns)
type UpdateProfileRequest struct {
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

// UserPost represents a simplified post for user profile display
type UserPost struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Title        *string   `json:"title,omitempty"`
	Content      string    `json:"content"`
	ImagePath    *string   `json:"image_path,omitempty"`
	PrivacyLevel string    `json:"privacy_level"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProfileResponse represents a comprehensive user profile with activity
type ProfileResponse struct {
	User           *User      `json:"user"`
	Posts          []UserPost `json:"posts"`
	Followers      []User     `json:"followers"`
	Following      []User     `json:"following"`
	FollowerCount  int        `json:"follower_count"`
	FollowingCount int        `json:"following_count"`
	PostCount      int        `json:"post_count"`
	CanView        bool       `json:"can_view"` // Whether viewer has access to this profile
}
