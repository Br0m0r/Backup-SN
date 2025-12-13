package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social-network/services/posts/middleware"
	"social-network/services/posts/utils"
)

const (
	maxUploadSize = 5 << 20 // 5MB
	uploadDir     = "./uploads/posts"
)

// UploadHandlers handles file upload requests
type UploadHandlers struct{}

// NewUploadHandlers creates a new upload handlers instance
func NewUploadHandlers() *UploadHandlers {
	// Ensure upload directory exists
	os.MkdirAll(uploadDir, os.ModePerm)
	return &UploadHandlers{}
}

// UploadImage handles POST /upload/image requests
func (h *UploadHandlers) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		if err.Error() == "http: request body too large" {
			utils.ErrorResponse(w, "File too large (max 5MB)", http.StatusBadRequest)
		} else {
			utils.ErrorResponse(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		}
		return
	}

	// Get the file from form
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		utils.ErrorResponse(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		utils.ErrorResponse(w, "Invalid file type. Only JPG, PNG, and GIF allowed", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d%s", userID, timestamp, ext)
	filePath := filepath.Join(uploadDir, filename)

	// Create the file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Return the file path (relative, without leading slash)
	relativePath := fmt.Sprintf("uploads/posts/%s", filename)
	utils.SuccessResponse(w, map[string]string{
		"image_path": relativePath,
		"filename":   filename,
	})
}

// DeleteImage handles DELETE /upload/image requests
func (h *UploadHandlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get image path from query parameter
	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		utils.ErrorResponse(w, "image path required", http.StatusBadRequest)
		return
	}

	// Extract filename from path
	filename := filepath.Base(imagePath)
	fullPath := filepath.Join(uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.ErrorResponse(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		utils.ErrorResponse(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]string{
		"message": "Image deleted successfully",
	})
}
