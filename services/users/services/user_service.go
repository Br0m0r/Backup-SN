package services

import (
	"database/sql"
	"errors"

	"social-network/services/users/db"
	"social-network/services/users/models"
)

// UserService handles user-related business logic
type UserService struct {
	database *sql.DB
}

// NewUserService creates a new user service instance
func NewUserService(database *sql.DB) *UserService {
	return &UserService{
		database: database,
	}
}

// GetProfile retrieves a user's profile
func (s *UserService) GetProfile(userID int) (*models.User, error) {
	return db.GetUserByID(s.database, userID)
}

// UpdateProfile updates a user's profile
func (s *UserService) UpdateProfile(userID int, req *models.UpdateProfileRequest) (*models.User, error) {
	err := db.UpdateUserProfile(s.database, userID, req)
	if err != nil {
		return nil, err
	}

	// Return updated profile
	return db.GetUserByID(s.database, userID)
}

// FollowUser creates a follow relationship
func (s *UserService) FollowUser(followerID, followingID int) error {
	// Check if trying to follow self
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	// Check if already following
	status, err := db.CheckFollowStatus(s.database, followerID, followingID)
	if err != nil {
		return err
	}

	if status == "accepted" || status == "pending" {
		return errors.New("already following or request pending")
	}

	// Get the user being followed to check if profile is public
	targetUser, err := db.GetUserByID(s.database, followingID)
	if err != nil {
		return errors.New("target user not found")
	}

	// If target profile is public, accept immediately; otherwise, set to pending
	followStatus := "accepted"
	if !targetUser.IsPublicProfile {
		followStatus = "pending"
	}

	return db.CreateFollow(s.database, followerID, followingID, followStatus)
}

// UnfollowUser removes a follow relationship
func (s *UserService) UnfollowUser(followerID, followingID int) error {
	return db.DeleteFollow(s.database, followerID, followingID)
}

// GetFollowers retrieves all followers of a user
func (s *UserService) GetFollowers(userID int) ([]*models.User, error) {
	return db.GetFollowers(s.database, userID)
}

// GetFollowing retrieves all users that a user is following
func (s *UserService) GetFollowing(userID int) ([]*models.User, error) {
	return db.GetFollowing(s.database, userID)
}

// SearchUsers searches for users
func (s *UserService) SearchUsers(searchTerm string) ([]*models.User, error) {
	if searchTerm == "" {
		return nil, errors.New("search term cannot be empty")
	}

	return db.SearchUsers(s.database, searchTerm)
}
