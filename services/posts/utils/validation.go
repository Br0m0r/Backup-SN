package utils

import (
	"errors"
	"html"
	"net/url"
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

	imageURL := *imagePath
	if imageURL != strings.TrimSpace(imageURL) || strings.Contains(imageURL, `\`) {
		return errors.New("Invalid image path")
	}

	parsed, err := url.Parse(imageURL)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.User != nil {
		return errors.New("Image path must use the local media route")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("Invalid image path")
	}
	decodedPath, err := url.PathUnescape(parsed.EscapedPath())
	if err != nil {
		return errors.New("Invalid image path")
	}

	// Check for path traversal attempts
	if strings.Contains(decodedPath, "..") {
		return errors.New("Invalid image path")
	}

	// Object-storage uploads are returned through the gateway's same-origin
	// media route. Other absolute paths remain invalid; relative paths are
	// temporarily accepted for legacy records until all media is migrated.
	if strings.HasPrefix(decodedPath, "/") && !strings.HasPrefix(decodedPath, "/media/") {
		return errors.New("Image path must use the local media route")
	}
	if decodedPath == "/media/" {
		return errors.New("Invalid image path")
	}

	return nil
}
