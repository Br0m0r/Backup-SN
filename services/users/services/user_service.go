package services

import (
	"database/sql"
	"errors"

	"social-network/services/common/notify"
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

	err = db.CreateFollow(s.database, followerID, followingID, followStatus)
	if err != nil {
		return err
	}

	// Send notification
	follower, _ := db.GetUserByID(s.database, followerID)
	if followStatus == "pending" {
		// Private profile - notify about follow request
		notify.FollowRequest(followingID, followerID, follower.Username)
	} else {
		// Public profile - notify about new follower
		notify.NewFollower(followingID, followerID, follower.Username)
	}

	return nil
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

// GetFollowStatus checks the follow relationship status between two users
func (s *UserService) GetFollowStatus(followerID, followingID int) (string, error) {
	return db.CheckFollowStatus(s.database, followerID, followingID)
}

// GetPendingFollowRequests retrieves all pending follow requests for a user
func (s *UserService) GetPendingFollowRequests(userID int) ([]*models.User, error) {
	return db.GetPendingFollowRequests(s.database, userID)
}

// RespondToFollowRequest accepts or rejects a follow request
func (s *UserService) RespondToFollowRequest(followerID, followingID int, accept bool) error {
	err := db.RespondToFollowRequest(s.database, followerID, followingID, accept)
	if err != nil {
		return err
	}

	// Send notification if accepted
	if accept {
		accepter, _ := db.GetUserByID(s.database, followingID)
		notify.FollowAccepted(followerID, followingID, accepter.Username)
	}

	return nil
}

// GetUserProfile retrieves a comprehensive user profile with posts, followers, and following
// Respects privacy settings: only returns full data if viewer has access
func (s *UserService) GetUserProfile(userID, viewerID int) (*models.ProfileResponse, error) {
	// Check if viewer can access this profile
	canView, err := db.CheckProfileAccess(s.database, userID, viewerID)
	if err != nil {
		return nil, err
	}

	// Get basic user info
	user, err := db.GetUserByID(s.database, userID)
	if err != nil {
		return nil, err
	}

	// If viewer cannot access profile, return limited info
	if !canView {
		return &models.ProfileResponse{
			User:           user.PublicProfile(),
			Posts:          []models.UserPost{},
			Followers:      []models.User{},
			Following:      []models.User{},
			FollowerCount:  0,
			FollowingCount: 0,
			PostCount:      0,
			CanView:        false,
		}, nil
	}

	// Get user's posts
	posts, err := db.GetUserPosts(s.database, userID)
	if err != nil {
		posts = []models.UserPost{} // If error, return empty slice
	}

	// Get followers
	followers, err := db.GetUserFollowersList(s.database, userID)
	if err != nil {
		followers = []models.User{}
	}

	// Get following
	following, err := db.GetUserFollowingList(s.database, userID)
	if err != nil {
		following = []models.User{}
	}

	// If viewer is not the owner, show public profiles only for followers/following
	var publicFollowers []models.User
	var publicFollowing []models.User

	if userID != viewerID {
		// Return public profiles for privacy
		for _, follower := range followers {
			publicFollowers = append(publicFollowers, *follower.PublicProfile())
		}
		for _, follow := range following {
			publicFollowing = append(publicFollowing, *follow.PublicProfile())
		}
	} else {
		// Owner sees everything
		publicFollowers = followers
		publicFollowing = following
	}

	// Build and return comprehensive profile
	profile := &models.ProfileResponse{
		User:           user,
		Posts:          posts,
		Followers:      publicFollowers,
		Following:      publicFollowing,
		FollowerCount:  len(followers),
		FollowingCount: len(following),
		PostCount:      len(posts),
		CanView:        true,
	}

	// If viewer is not owner, hide sensitive user info
	if userID != viewerID {
		profile.User = user.PublicProfile()
	}

	return profile, nil
}
