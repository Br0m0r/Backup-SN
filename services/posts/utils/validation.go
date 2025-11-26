package utils

import (
	"errors"
	"html"
	"regexp"
	"strings"
)

const (
	MaxPostContentLength = 1000
	MaxTitleLength       = 200
	MinPostContentLength = 1
)

var dangerousRegex = regexp.MustCompile(`(?i)<script|javascript:|onerror=|onload=|<iframe|eval\(|<embed|<object`)

// ValidatePostContent validates and sanitizes post content
func ValidatePostContent(content string, allowEmpty bool) (string, error) {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Check if empty
	if content == "" && !allowEmpty {
		return "", errors.New("Post content cannot be empty")
	}

	if content == "" {
		return "", nil
	}

	// Check length
	if len(content) < MinPostContentLength {
		return "", errors.New("Post content is too short")
	}
	if len(content) > MaxPostContentLength {
		return "", errors.New("Post content is too long (max 10000 characters)")
	}

	// Check for dangerous patterns (XSS attempts)
	if dangerousRegex.MatchString(content) {
		return "", errors.New("Post content contains potentially dangerous code")
	}

	// Escape HTML to prevent XSS
	sanitized := html.EscapeString(content)

	return sanitized, nil
}

// ValidateTitle validates and sanitizes post title
func ValidateTitle(title *string) (*string, error) {
	if title == nil {
		return nil, nil
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(*title)

	if trimmed == "" {
		return nil, nil // Empty title is allowed
	}

	// Check length
	if len(trimmed) > MaxTitleLength {
		return nil, errors.New("Post title is too long (max 200 characters)")
	}

	// Check for dangerous patterns
	if dangerousRegex.MatchString(trimmed) {
		return nil, errors.New("Post title contains potentially dangerous code")
	}

	// Escape HTML
	sanitized := html.EscapeString(trimmed)

	return &sanitized, nil
}

// ValidateImagePath validates image path to prevent path traversal
func ValidateImagePath(imagePath *string) error {
	if imagePath == nil || *imagePath == "" {
		return nil
	}

	path := *imagePath

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return errors.New("Invalid image path")
	}

	// Check for absolute paths
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return errors.New("Image path must be relative")
	}

	return nil
}
