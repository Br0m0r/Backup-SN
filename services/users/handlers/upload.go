package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"social-network/services/common/objectstore"
	"social-network/services/users/middleware"
	"social-network/services/users/services"
	"social-network/services/users/utils"
)

const maxAvatarUploadSize = 5 << 20 // 5 MiB

// UploadHandlers handles avatar object storage operations.
type UploadHandlers struct {
	userService *services.UserService
	store       objectstore.Store
}

func NewUploadHandlers(userService *services.UserService, store objectstore.Store) *UploadHandlers {
	return &UploadHandlers{userService: userService, store: store}
}

// UploadAvatar handles POST /upload/avatar.
func (h *UploadHandlers) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadSize+(1<<20))
	if err := r.ParseMultipartForm(maxAvatarUploadSize); err != nil {
		utils.ErrorResponse(w, "File too large or invalid form data (max 5MB)", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contents, err := io.ReadAll(io.LimitReader(file, maxAvatarUploadSize+1))
	if err != nil || len(contents) == 0 || len(contents) > maxAvatarUploadSize {
		utils.ErrorResponse(w, "Invalid avatar or file too large (max 5MB)", http.StatusBadRequest)
		return
	}
	contentType := http.DetectContentType(contents)
	extension, ok := avatarImageExtension(contentType)
	if !ok {
		utils.ErrorResponse(w, "Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed", http.StatusBadRequest)
		return
	}

	objectName, err := randomAvatarObjectName()
	if err != nil {
		log.Printf("Failed to generate avatar media key: %v", err)
		utils.ErrorResponse(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("avatars/users/%d/%s%s", userID, objectName, extension)
	if err := h.store.Put(r.Context(), key, bytes.NewReader(contents), int64(len(contents)), contentType); err != nil {
		log.Printf("Failed to upload avatar media: %v", err)
		utils.ErrorResponse(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}

	avatarPath := h.store.URL(key)
	previous, _ := h.userService.GetProfile(userID)
	if err := h.userService.UpdateUserAvatarPath(userID, avatarPath); err != nil {
		_ = h.store.Delete(r.Context(), key)
		log.Printf("Failed to update user avatar path: %v", err)
		utils.ErrorResponse(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	}

	// Once the new reference is durable, remove the previous owned object.
	if previous != nil && previous.AvatarPath != nil && *previous.AvatarPath != "" {
		if previousKey, err := h.store.KeyFromURL(*previous.AvatarPath); err == nil && strings.HasPrefix(previousKey, fmt.Sprintf("avatars/users/%d/", userID)) {
			if err := h.store.Delete(r.Context(), previousKey); err != nil {
				log.Printf("Failed to remove superseded avatar %q: %v", previousKey, err)
			}
		}
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"avatar_path": avatarPath,
		"message":     "Avatar uploaded successfully",
	})
}

// DeleteAvatar handles DELETE /upload/avatar.
func (h *UploadHandlers) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	avatarPath := r.URL.Query().Get("path")
	if avatarPath == "" {
		utils.ErrorResponse(w, "Avatar path required", http.StatusBadRequest)
		return
	}

	key, err := h.store.KeyFromURL(avatarPath)
	if err != nil || !strings.HasPrefix(key, fmt.Sprintf("avatars/users/%d/", userID)) {
		utils.ErrorResponse(w, "Cannot delete another user's avatar", http.StatusForbidden)
		return
	}
	if err := h.userService.ClearUserAvatarPath(userID, avatarPath); err != nil {
		log.Printf("Failed to clear user avatar path: %v", err)
		utils.ErrorResponse(w, "Failed to delete avatar", http.StatusInternalServerError)
		return
	}
	if err := h.store.Delete(r.Context(), key); err != nil {
		log.Printf("Failed to delete avatar object: %v", err)
		utils.ErrorResponse(w, "Failed to delete avatar", http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{"message": "Avatar deleted successfully"})
}

func randomAvatarObjectName() (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(random[:]), nil
}

func avatarImageExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
