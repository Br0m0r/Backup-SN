package models

import "time"

// Post represents a user post
type Post struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Title        *string   `json:"title,omitempty"`
	Content      string    `json:"content"`
	ImagePath    *string   `json:"image_path,omitempty"`
	PrivacyLevel string    `json:"privacy_level"` // "public", "private", "almost_private"
	CreatedAt    time.Time `json:"created_at"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	ImagePath *string   `json:"image_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PostViewer represents a user who can view an "almost_private" post
type PostViewer struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	Title        *string `json:"title,omitempty"`
	Content      string  `json:"content"`
	ImagePath    *string `json:"image_path,omitempty"`
	PrivacyLevel string  `json:"privacy_level"`     // "public", "private", "almost_private"
	Viewers      []int   `json:"viewers,omitempty"` // User IDs for "almost_private" posts
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Title        *string `json:"title,omitempty"`
	Content      string  `json:"content"`
	ImagePath    *string `json:"image_path,omitempty"`
	PrivacyLevel string  `json:"privacy_level"`
	Viewers      []int   `json:"viewers,omitempty"` // User IDs for "almost_private" posts
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	PostID    int     `json:"post_id"`
	Content   string  `json:"content"`
	ImagePath *string `json:"image_path,omitempty"`
}
