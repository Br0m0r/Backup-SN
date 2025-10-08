package utils

import (
	"errors"
	"regexp"
	"strings"

	"social-network/services/auth/models"
)

// ValidateRegisterRequest validates the registration request
func ValidateRegisterRequest(req *models.RegisterRequest) error {
	// Username validation
	if req.Username == "" {
		return errors.New("username is required")
	}

	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	// Email validation
	if req.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}

	// Password validation
	if req.Password == "" {
		return errors.New("password is required")
	}

	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	// Name validation
	if strings.TrimSpace(req.FirstName) == "" {
		return errors.New("first name is required")
	}

	if strings.TrimSpace(req.LastName) == "" {
		return errors.New("last name is required")
	}

	return nil
}

// ValidateLoginRequest validates the login request
func ValidateLoginRequest(req *models.LoginRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	return nil
}
