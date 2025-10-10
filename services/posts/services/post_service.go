package services

import (
	"database/sql"
	"errors"
	"time"

	"social-network/services/posts/db"
	"social-network/services/posts/models"
)

// PostService handles business logic for posts
type PostService struct {
	database *sql.DB
}

// NewPostService creates a new post service instance
func NewPostService(database *sql.DB) *PostService {
	return &PostService{
		database: database,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(req *models.CreatePostRequest, userID int) (*models.Post, error) {
	// Validate privacy level
	if req.PrivacyLevel != "public" && req.PrivacyLevel != "private" && req.PrivacyLevel != "almost_private" {
		return nil, errors.New("invalid privacy level")
	}

	// Validate content
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	// Create post
	post := &models.Post{
		UserID:       userID,
		Content:      req.Content,
		ImagePath:    req.ImagePath,
		PrivacyLevel: req.PrivacyLevel,
		CreatedAt:    time.Now(),
	}

	err := db.CreatePost(s.database, post)
	if err != nil {
		return nil, err
	}

	// Add viewers if almost_private
	if req.PrivacyLevel == "almost_private" && len(req.Viewers) > 0 {
		err = db.AddPostViewers(s.database, post.ID, req.Viewers)
		if err != nil {
			return nil, err
		}
	}

	return post, nil
}

// GetPost retrieves a post by ID with access check
func (s *PostService) GetPost(postID, userID int) (*models.Post, error) {
	// Check if user has access to the post
	hasAccess, err := db.CheckPostAccess(s.database, postID, userID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("access denied")
	}

	// Get post
	post, err := db.GetPostByID(s.database, postID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(postID, userID int, req *models.UpdatePostRequest) (*models.Post, error) {
	// Get existing post
	post, err := db.GetPostByID(s.database, postID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if post.UserID != userID {
		return nil, errors.New("unauthorized: you can only update your own posts")
	}

	// Validate privacy level
	if req.PrivacyLevel != "public" && req.PrivacyLevel != "private" && req.PrivacyLevel != "almost_private" {
		return nil, errors.New("invalid privacy level")
	}

	// Validate content
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	// Update post fields
	post.Content = req.Content
	post.ImagePath = req.ImagePath
	post.PrivacyLevel = req.PrivacyLevel

	err = db.UpdatePost(s.database, post)
	if err != nil {
		return nil, err
	}

	// Update viewers if almost_private
	if req.PrivacyLevel == "almost_private" {
		err = db.AddPostViewers(s.database, post.ID, req.Viewers)
		if err != nil {
			return nil, err
		}
	} else {
		// Clear viewers if privacy level changed
		err = db.AddPostViewers(s.database, post.ID, []int{})
		if err != nil {
			return nil, err
		}
	}

	return post, nil
}

// DeletePost deletes a post
func (s *PostService) DeletePost(postID, userID int) error {
	// Get post
	post, err := db.GetPostByID(s.database, postID)
	if err != nil {
		return err
	}

	// Check ownership
	if post.UserID != userID {
		return errors.New("unauthorized: you can only delete your own posts")
	}

	// Delete post (cascade will delete comments and viewers)
	return db.DeletePost(s.database, postID)
}

// GetFeed retrieves posts for a user's feed
func (s *PostService) GetFeed(userID int) ([]*models.Post, error) {
	return db.GetFeedPosts(s.database, userID)
}

// CreateComment creates a new comment on a post
func (s *PostService) CreateComment(req *models.CreateCommentRequest, userID int) (*models.Comment, error) {
	// Check if user has access to the post
	hasAccess, err := db.CheckPostAccess(s.database, req.PostID, userID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("access denied: cannot comment on this post")
	}

	// Validate content
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	// Create comment
	comment := &models.Comment{
		PostID:    req.PostID,
		UserID:    userID,
		Content:   req.Content,
		ImagePath: req.ImagePath,
		CreatedAt: time.Now(),
	}

	err = db.CreateComment(s.database, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComments retrieves comments for a post (with access check)
func (s *PostService) GetComments(postID, userID int) ([]*models.Comment, error) {
	// Check if user has access to the post
	hasAccess, err := db.CheckPostAccess(s.database, postID, userID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("access denied: cannot view comments on this post")
	}

	// Get comments
	return db.GetCommentsByPostID(s.database, postID)
}
