package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"social-network/services/common/notify"
	"social-network/services/posts/db"
	"social-network/services/posts/models"
	"social-network/services/posts/usersclient"
	"social-network/services/posts/utils"
)

// PostService handles business logic for posts
type PostService struct {
	database  *sql.DB
	directory usersclient.Directory
}

// NewPostService creates a new post service instance
func NewPostService(database *sql.DB, directory usersclient.Directory) *PostService {
	return &PostService{
		database:  database,
		directory: directory,
	}
}

func (s *PostService) Ping(ctx context.Context) error {
	return s.database.PingContext(ctx)
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
	hasAccess, err := s.canAccess(postID, userID)
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

	if err := s.hydratePosts([]*models.Post{post}); err != nil {
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

	if err := s.hydratePosts([]*models.Post{post}); err != nil {
		return nil, err
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
	followingIDs, err := s.directory.FollowingIDs(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	posts, err := db.GetFeedPosts(s.database, userID, followingIDs)
	if err != nil {
		return nil, err
	}
	if err := s.hydratePosts(posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// SearchPosts searches for posts based on query string
func (s *PostService) SearchPosts(userID int, query string) ([]*models.Post, error) {
	if query == "" {
		return []*models.Post{}, nil
	}
	followingIDs, err := s.directory.FollowingIDs(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	matchingProfiles, err := s.directory.SearchProfiles(context.Background(), query)
	if err != nil {
		return nil, err
	}
	authorIDs := make([]int, 0, len(matchingProfiles))
	for _, profile := range matchingProfiles {
		authorIDs = append(authorIDs, profile.ID)
	}
	posts, err := db.SearchPosts(s.database, userID, query, followingIDs, authorIDs)
	if err != nil {
		return nil, err
	}
	if err := s.hydratePosts(posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// CreateComment creates a new comment on a post
func (s *PostService) CreateComment(req *models.CreateCommentRequest, userID int, commenterName string) (*models.Comment, error) {
	// Check if user has access to the post
	hasAccess, err := s.canAccess(req.PostID, userID)
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
	hasAccess, err := s.canAccess(postID, userID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("access denied: cannot view comments on this post")
	}

	// Get comments
	comments, err := db.GetCommentsByPostID(s.database, postID)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateComments(comments); err != nil {
		return nil, err
	}
	return comments, nil
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

	if err := s.hydrateComments([]*models.Comment{comment}); err != nil {
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
	posts, err := db.GetPostsByGroupID(s.database, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.hydratePosts(posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *PostService) GetUserPosts(userID int) ([]*models.Post, error) {
	return db.GetPostsByUserID(s.database, userID)
}

func (s *PostService) canAccess(postID, userID int) (bool, error) {
	post, err := db.GetPostByID(s.database, postID)
	if err != nil {
		return false, err
	}
	followingIDs, err := s.directory.FollowingIDs(context.Background(), userID)
	if err != nil {
		return false, err
	}
	followsAuthor := false
	for _, followingID := range followingIDs {
		if followingID == post.UserID {
			followsAuthor = true
			break
		}
	}
	return db.CheckPostAccess(s.database, postID, userID, followsAuthor)
}

func (s *PostService) hydratePosts(posts []*models.Post) error {
	userIDs := uniquePostAuthorIDs(posts)
	profiles, err := s.directory.Profiles(context.Background(), userIDs)
	if err != nil {
		return err
	}
	authors := authorsByID(profiles)
	for _, post := range posts {
		post.Author = authors[post.UserID]
	}
	return nil
}

func (s *PostService) hydrateComments(comments []*models.Comment) error {
	seen := make(map[int]struct{})
	userIDs := make([]int, 0)
	for _, comment := range comments {
		if _, exists := seen[comment.UserID]; !exists {
			seen[comment.UserID] = struct{}{}
			userIDs = append(userIDs, comment.UserID)
		}
	}
	profiles, err := s.directory.Profiles(context.Background(), userIDs)
	if err != nil {
		return err
	}
	authors := authorsByID(profiles)
	for _, comment := range comments {
		comment.Author = authors[comment.UserID]
	}
	return nil
}

func uniquePostAuthorIDs(posts []*models.Post) []int {
	seen := make(map[int]struct{})
	userIDs := make([]int, 0)
	for _, post := range posts {
		if _, exists := seen[post.UserID]; !exists {
			seen[post.UserID] = struct{}{}
			userIDs = append(userIDs, post.UserID)
		}
	}
	return userIDs
}

func authorsByID(profiles []usersclient.Profile) map[int]*models.Author {
	authors := make(map[int]*models.Author, len(profiles))
	for _, profile := range profiles {
		author := &models.Author{ID: profile.ID, Username: profile.Username}
		if profile.FirstName != nil {
			author.FirstName = *profile.FirstName
		}
		if profile.LastName != nil {
			author.LastName = *profile.LastName
		}
		if profile.AvatarPath != nil {
			author.AvatarPath = *profile.AvatarPath
		}
		authors[profile.ID] = author
	}
	return authors
}
