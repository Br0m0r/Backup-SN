package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"social-network/services/auth/db"
	"social-network/services/auth/models"
	"social-network/services/auth/usersclient"
	"social-network/services/auth/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	database     *sql.DB
	tokenService *TokenService
	profiles     usersclient.Provisioner
}

// NewAuthService creates a new auth service instance
func NewAuthService(database *sql.DB, profiles usersclient.Provisioner) *AuthService {
	return &AuthService{
		database:     database,
		tokenService: NewTokenService(database),
		profiles:     profiles,
	}
}

func (s *AuthService) Ping(ctx context.Context) error {
	return s.database.PingContext(ctx)
}

// Register creates a new user account
func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate request
	if err := utils.ValidateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Generate username from email if not provided
	if req.Username == "" {
		// Use the part before @ in email as username
		emailParts := strings.Split(req.Email, "@")
		req.Username = emailParts[0]
	}

	// Check if username already exists
	if exists, err := db.UserExistsByUsername(s.database, req.Username); err != nil {
		return nil, errors.New("database error checking username existence")
	} else if exists {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if exists, err := db.UserExistsByEmail(s.database, req.Email); err != nil {
		return nil, errors.New("database error checking email existence")
	} else if exists {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user in database
	user, err := db.CreateUser(s.database, req.Username, req.Email, hashedPassword, req.FirstName, req.LastName, req.DateOfBirth, req.Nickname, req.AboutMe)
	if err != nil {
		return nil, err
	}
	if s.profiles != nil {
		err = s.profiles.Provision(context.Background(), usersclient.Profile{
			AccountID: user.ID, Username: user.Username, Email: user.Email,
			FirstName: req.FirstName, LastName: req.LastName,
			DateOfBirth: req.DateOfBirth, Nickname: req.Nickname, AboutMe: req.AboutMe,
		})
		if err != nil {
			_ = db.DeleteUser(s.database, user.ID)
			return nil, errors.New("failed to provision user profile")
		}
	}

	// Generate token
	token, err := s.tokenService.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate authentication token")
	}

	return &models.AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Validate request
	if err := utils.ValidateLoginRequest(req); err != nil {
		return nil, err
	}

	// Get user from database
	user, err := db.GetUserByEmail(s.database, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Generate token
	token, err := s.tokenService.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate authentication token")
	}

	return &models.AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// VerifyToken validates an authentication token
func (s *AuthService) VerifyToken(token string) (*models.User, error) {
	// Validate token
	sessionData, err := s.tokenService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Get updated user data from database
	user, err := db.GetUserByID(s.database, sessionData.UserID)
	if err != nil {
		// If user no longer exists, invalidate the token
		s.tokenService.InvalidateToken(token)
		return nil, errors.New("user not found")
	}

	return user, nil
}

// Logout invalidates a user's token
func (s *AuthService) Logout(token string) error {
	return s.tokenService.InvalidateToken(token)
}

// GetUserByID retrieves user information by ID (for internal service communication)
func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	user, err := db.GetUserByID(s.database, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
