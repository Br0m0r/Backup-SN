package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social-network/services/users/middleware"
	"social-network/services/users/services"
	"social-network/services/users/utils"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
	uploadDir     = "./uploads/avatars"
)

// UploadHandlers handles file upload operations
type UploadHandlers struct {
	userService *services.UserService
}

// NewUploadHandlers creates a new upload handlers instance
func NewUploadHandlers(userService *services.UserService) *UploadHandlers {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Warning: Failed to create upload directory: %v", err)
	}
	return &UploadHandlers{
		userService: userService,
	}
}

// UploadAvatar handles POST /upload/avatar
func (h *UploadHandlers) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		utils.ErrorResponse(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		utils.ErrorResponse(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		utils.ErrorResponse(w, "Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromContentType(contentType)
	}
	filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error saving file: %v", err)
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Update user's avatar_path in database
	avatarPath := fmt.Sprintf("/uploads/avatars/%s", filename)
	if err := h.userService.UpdateUserAvatarPath(userID, avatarPath); err != nil {
		log.Printf("Error updating user avatar path: %v", err)
		// File was saved but DB update failed - log but don't fail the request
		// The avatar is still accessible via the URL
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"avatar_path": avatarPath,
		"message":     "Avatar uploaded successfully",
	})
}

// DeleteAvatar handles DELETE /upload/avatar
func (h *UploadHandlers) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get avatar path from query parameter
	avatarPath := r.URL.Query().Get("path")
	if avatarPath == "" {
		utils.ErrorResponse(w, "Avatar path required", http.StatusBadRequest)
		return
	}

	// Extract filename from path
	filename := filepath.Base(avatarPath)

	// Build full file path
	fullPath := filepath.Join(uploadDir, filename)

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			utils.ErrorResponse(w, "Avatar not found", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting avatar: %v", err)
		utils.ErrorResponse(w, "Failed to delete avatar", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Avatar deleted successfully",
	})
}

// Helper functions

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	for _, validType := range validTypes {
		if strings.EqualFold(contentType, validType) {
			return true
		}
	}
	return false
}

func getExtensionFromContentType(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
