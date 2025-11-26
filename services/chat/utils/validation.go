package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// Message limits
	MaxMessageLength   = 500 // 500 characters
	MaxImagePathLength = 200 // 200 characters for file paths
	MinMessageLength   = 1   // At least 1 character (if no image)

	// Patterns that might indicate XSS attempts
	dangerousPatterns = `<script|javascript:|onerror=|onload=|onclick=|<iframe|<object|<embed|eval\(|expression\(|vbscript:`
)

var (
	// Regex to detect dangerous patterns
	dangerousRegex = regexp.MustCompile(`(?i)` + dangerousPatterns)

	// Regex to detect excessive whitespace
	excessiveWhitespace = regexp.MustCompile(`\s{100,}`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateMessageContent validates and sanitizes message content
// Returns sanitized content or error
func ValidateMessageContent(content string, allowEmpty bool) (string, error) {
	// Check if empty (only if image is not provided)
	trimmed := strings.TrimSpace(content)
	if !allowEmpty && len(trimmed) == 0 {
		return "", &ValidationError{
			Field:   "content",
			Message: "Message content cannot be empty",
		}
	}

	// Allow empty if image is provided
	if allowEmpty && len(trimmed) == 0 {
		return "", nil
	}

	// Check minimum length
	if !allowEmpty && utf8.RuneCountInString(trimmed) < MinMessageLength {
		return "", &ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("Message must be at least %d character", MinMessageLength),
		}
	}

	// Check maximum length (use rune count for Unicode support)
	if utf8.RuneCountInString(content) > MaxMessageLength {
		return "", &ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("Message exceeds maximum length of %d characters", MaxMessageLength),
		}
	}

	// Check for dangerous patterns (XSS attempts)
	if dangerousRegex.MatchString(content) {
		return "", &ValidationError{
			Field:   "content",
			Message: "Message contains potentially dangerous content",
		}
	}

	// Check for excessive whitespace (spam/DoS attempt)
	if excessiveWhitespace.MatchString(content) {
		return "", &ValidationError{
			Field:   "content",
			Message: "Message contains excessive whitespace",
		}
	}

	// Sanitize HTML entities to prevent XSS
	// This converts < to &lt;, > to &gt;, etc.
	sanitized := html.EscapeString(content)

	// Additional: Remove any null bytes (can cause issues)
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	return sanitized, nil
}

// ValidateImagePath validates image file paths
func ValidateImagePath(path string) error {
	if path == "" {
		return nil // Empty is OK
	}

	// Check length
	if len(path) > MaxImagePathLength {
		return &ValidationError{
			Field:   "image_path",
			Message: fmt.Sprintf("Image path exceeds maximum length of %d", MaxImagePathLength),
		}
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return &ValidationError{
			Field:   "image_path",
			Message: "Invalid image path",
		}
	}

	// Must start with / or uploads/ (your upload directory)
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "uploads/") {
		return &ValidationError{
			Field:   "image_path",
			Message: "Invalid image path format",
		}
	}

	return nil
}

// SanitizeForDisplay prepares content for safe display
// Use this when sending to frontend
func SanitizeForDisplay(content string) string {
	// Already sanitized during validation, but double-check
	sanitized := html.EscapeString(content)

	// Trim excessive whitespace for display
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ValidateGroupMessage validates group message content
// Same rules as regular messages
func ValidateGroupMessage(content string) (string, error) {
	return ValidateMessageContent(content, false)
}
