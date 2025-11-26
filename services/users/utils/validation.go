package utils

import (
	"errors"
	"html"
	"regexp"
	"strings"
)

const (
	MaxNicknameLength = 50
	MaxAboutMeLength  = 500
	MaxNameLength     = 100
)

var dangerousRegex = regexp.MustCompile(`(?i)<script|javascript:|onerror=|onload=|<iframe|eval\(|<embed|<object`)

// ValidateNickname validates and sanitizes nickname
func ValidateNickname(nickname *string) (*string, error) {
	if nickname == nil {
		return nil, nil
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(*nickname)

	if trimmed == "" {
		return nil, nil
	}

	// Check length
	if len(trimmed) > MaxNicknameLength {
		return nil, errors.New("Nickname is too long (max 50 characters)")
	}

	// Check for dangerous patterns
	if dangerousRegex.MatchString(trimmed) {
		return nil, errors.New("Nickname contains invalid characters")
	}

	// Escape HTML
	sanitized := html.EscapeString(trimmed)

	return &sanitized, nil
}

// ValidateAboutMe validates and sanitizes about me text
func ValidateAboutMe(aboutMe *string) (*string, error) {
	if aboutMe == nil {
		return nil, nil
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(*aboutMe)

	if trimmed == "" {
		return nil, nil
	}

	// Check length
	if len(trimmed) > MaxAboutMeLength {
		return nil, errors.New("About me text is too long (max 500 characters)")
	}

	// Check for dangerous patterns
	if dangerousRegex.MatchString(trimmed) {
		return nil, errors.New("About me text contains invalid content")
	}

	// Escape HTML
	sanitized := html.EscapeString(trimmed)

	return &sanitized, nil
}

// ValidateName validates and sanitizes first/last name
func ValidateName(name *string) (*string, error) {
	if name == nil {
		return nil, nil
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(*name)

	if trimmed == "" {
		return nil, nil
	}

	// Check length
	if len(trimmed) > MaxNameLength {
		return nil, errors.New("Name is too long (max 100 characters)")
	}

	// Check for dangerous patterns
	if dangerousRegex.MatchString(trimmed) {
		return nil, errors.New("Name contains invalid characters")
	}

	// Escape HTML
	sanitized := html.EscapeString(trimmed)

	return &sanitized, nil
}
