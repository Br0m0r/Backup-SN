package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"social-network/services/common/objectstore"
	"social-network/services/posts/middleware"
	"social-network/services/posts/utils"
)

const maxUploadSize = 5 << 20 // 5MB

// UploadHandlers handles file upload requests
type UploadHandlers struct {
	store objectstore.Store
}

// NewUploadHandlers creates a new upload handlers instance
func NewUploadHandlers(store objectstore.Store) *UploadHandlers {
	return &UploadHandlers{store: store}
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
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+(1<<20))
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		utils.ErrorResponse(w, "File too large or invalid form data (max 5MB)", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	// Get the file from form
	file, _, err := r.FormFile("image")
	if err != nil {
		utils.ErrorResponse(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contents, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil {
		utils.ErrorResponse(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	if len(contents) == 0 || len(contents) > maxUploadSize {
		utils.ErrorResponse(w, "File too large (max 5MB)", http.StatusBadRequest)
		return
	}
	contentType := http.DetectContentType(contents)
	extension, ok := postImageExtension(contentType)
	if !ok {
		utils.ErrorResponse(w, "Invalid file type. Only JPG, PNG, and GIF allowed", http.StatusBadRequest)
		return
	}

	objectName, err := randomPostObjectName()
	if err != nil {
		log.Printf("Failed to generate post media key: %v", err)
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("posts/users/%d/%s%s", userID, objectName, extension)
	if err := h.store.Put(r.Context(), key, bytes.NewReader(contents), int64(len(contents)), contentType); err != nil {
		log.Printf("Failed to upload post media: %v", err)
		utils.ErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]string{
		"image_path": h.store.URL(key),
		"filename":   path.Base(key),
	})
}

// DeleteImage handles DELETE /upload/image requests
func (h *UploadHandlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user ID from context
	userID, ok := middleware.GetUserIDFromContext(r)
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

	key, err := h.store.KeyFromURL(imagePath)
	if err != nil {
		utils.ErrorResponse(w, "Invalid image path", http.StatusBadRequest)
		return
	}
	ownerPrefix := fmt.Sprintf("posts/users/%d/", userID)
	if !strings.HasPrefix(key, ownerPrefix) {
		utils.ErrorResponse(w, "Cannot delete another user's image", http.StatusForbidden)
		return
	}
	if err := h.store.Delete(r.Context(), key); err != nil {
		log.Printf("Failed to delete post media: %v", err)
		utils.ErrorResponse(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]string{
		"message": "Image deleted successfully",
	})
}

func randomPostObjectName() (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(random[:]), nil
}

func postImageExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}
