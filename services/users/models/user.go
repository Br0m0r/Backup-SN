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
	FollowStatus    *string   `json:"follow_status,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// PublicProfile returns a user profile without sensitive information
func (u *User) PublicProfile() *User {
	return &User{
		ID:       u.ID,
		Username: u.Username,
		// Email omitted for privacy
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		DateOfBirth:     u.DateOfBirth, // Include date of birth in public profile
		AvatarPath:      u.AvatarPath,
		Nickname:        u.Nickname,
		AboutMe:         u.AboutMe,
		IsPublicProfile: u.IsPublicProfile,
		CreatedAt:       u.CreatedAt,
	}
}

// UpdateProfileRequest represents profile update payload
type UpdateProfileRequest struct {
	FirstName       *string `json:"first_name,omitempty"`
	LastName        *string `json:"last_name,omitempty"`
	DateOfBirth     *string `json:"date_of_birth,omitempty"`
	AvatarPath      *string `json:"avatar_path,omitempty"`
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

// ProfileSummary is the non-sensitive identity view shared with trusted
// services that render user-authored content.
type ProfileSummary struct {
	ID         int     `json:"id"`
	Username   string  `json:"username"`
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	AvatarPath *string `json:"avatar_path,omitempty"`
	Nickname   *string `json:"nickname,omitempty"`
}

// ChatContact is the Users-owned identity and relationship view consumed by
// Chat when building the available-contact list.
type ChatContact struct {
	ProfileSummary
	IsMessageRequest bool `json:"is_message_request"`
}

// ProvisionProfileRequest is the authenticated Auth-to-Users registration
// contract. AccountID is the stable identity allocated by Auth.
type ProvisionProfileRequest struct {
	AccountID   int     `json:"account_id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	DateOfBirth string  `json:"date_of_birth"`
	Nickname    *string `json:"nickname"`
	AboutMe     *string `json:"about_me"`
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
