package services

import (
	"database/sql"
	"errors"
	"time"

	"social-network/services/common/notify"
	"social-network/services/posts/db"
	"social-network/services/posts/models"
	"social-network/services/posts/utils"
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

	// Validate and sanitize content
	sanitizedContent, err := utils.ValidatePostContent(req.Content, false)
	if err != nil {
		return nil, err
	}

	// Validate and sanitize title
	sanitizedTitle, err := utils.ValidateTitle(req.Title)
	if err != nil {
		return nil, err
	}

	// Validate image path
	if err := utils.ValidateImagePath(req.ImagePath); err != nil {
		return nil, err
	}

	// Create post with sanitized content
	post := &models.Post{
		UserID:       userID,
		GroupID:      req.GroupID,
		Title:        sanitizedTitle,
		Content:      sanitizedContent,
		ImagePath:    req.ImagePath,
		PrivacyLevel: req.PrivacyLevel,
		CreatedAt:    time.Now(),
	}

	err = db.CreatePost(s.database, post)
	if err != nil {
		return nil, err
	}

	// Add viewers if private (specific chosen followers)
	if req.PrivacyLevel == "private" && len(req.Viewers) > 0 {
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

	// Update viewers if private (specific chosen followers)
	if req.PrivacyLevel == "private" {
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

// SearchPosts searches for posts based on query string
func (s *PostService) SearchPosts(userID int, query string) ([]*models.Post, error) {
	if query == "" {
		return []*models.Post{}, nil
	}
	return db.SearchPosts(s.database, userID, query)
}

// CreateComment creates a new comment on a post
func (s *PostService) CreateComment(req *models.CreateCommentRequest, userID int, commenterName string) (*models.Comment, error) {
	// Check if user has access to the post
	hasAccess, err := db.CheckPostAccess(s.database, req.PostID, userID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("access denied: cannot comment on this post")
	}

	// Validate and sanitize content
	sanitizedContent, err := utils.ValidatePostContent(req.Content, false)
	if err != nil {
		return nil, err
	}

	// Validate image path
	if err := utils.ValidateImagePath(req.ImagePath); err != nil {
		return nil, err
	}

	// Create comment with sanitized content
	comment := &models.Comment{
		PostID:    req.PostID,
		UserID:    userID,
		Content:   sanitizedContent,
		ImagePath: req.ImagePath,
		CreatedAt: time.Now(),
	}

	err = db.CreateComment(s.database, comment)
	if err != nil {
		return nil, err
	}

	// Get post author and send notification (don't notify if commenting on own post)
	post, err := db.GetPostByID(s.database, req.PostID)
	if err == nil && post.UserID != userID {
		// Truncate content for preview
		preview := sanitizedContent
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		notify.NewComment(post.UserID, comment.ID, commenterName, preview)
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

// UpdateComment updates an existing comment
func (s *PostService) UpdateComment(commentID, userID int, content string, imagePath *string) (*models.Comment, error) {
	// Get existing comment
	comment, err := db.GetCommentByID(s.database, commentID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if comment.UserID != userID {
		return nil, errors.New("unauthorized: you can only update your own comments")
	}

	// Validate and sanitize content
	sanitizedContent, err := utils.ValidatePostContent(content, false)
	if err != nil {
		return nil, err
	}

	// Validate image path
	if err := utils.ValidateImagePath(imagePath); err != nil {
		return nil, err
	}

	// Update comment
	comment.Content = sanitizedContent
	comment.ImagePath = imagePath

	err = db.UpdateComment(s.database, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment deletes a comment
func (s *PostService) DeleteComment(commentID, userID int) error {
	// Get comment
	comment, err := db.GetCommentByID(s.database, commentID)
	if err != nil {
		return err
	}

	// Check ownership
	if comment.UserID != userID {
		return errors.New("unauthorized: you can only delete your own comments")
	}

	// Delete comment
	return db.DeleteComment(s.database, commentID)
}

// GetGroupPosts retrieves all posts for a specific group
func (s *PostService) GetGroupPosts(groupID int) ([]*models.Post, error) {
	return db.GetPostsByGroupID(s.database, groupID)
}
