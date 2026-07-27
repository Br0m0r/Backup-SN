package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"social-network/services/chat/middleware"
)

type fakeMediaStore struct {
	putKey         string
	putBody        []byte
	putContentType string
	deletedKey     string
	keyFromURL     string
	putErr         error
	deleteErr      error
}

func (store *fakeMediaStore) Put(_ context.Context, key string, reader io.Reader, _ int64, contentType string) error {
	store.putKey = key
	store.putContentType = contentType
	store.putBody, _ = io.ReadAll(reader)
	return store.putErr
}

func (store *fakeMediaStore) Delete(_ context.Context, key string) error {
	store.deletedKey = key
	return store.deleteErr
}

func (store *fakeMediaStore) URL(key string) string {
	return "/media/social-network-media/" + key
}

func (store *fakeMediaStore) KeyFromURL(string) (string, error) {
	return store.keyFromURL, nil
}

func TestUploadImageStoresDetectedPNGWithOpaqueName(t *testing.T) {
	image := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 32)...)
	request := multipartRequest(t, "malicious-name.jpg", image)
	request = middleware.SetUserIDInContext(request, 7)
	recorder := httptest.NewRecorder()
	store := &fakeMediaStore{}

	NewUploadHandlers(store).UploadImage(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.HasPrefix(store.putKey, "chat/users/7/") || !strings.HasSuffix(store.putKey, ".png") {
		t.Fatalf("unexpected object key: %q", store.putKey)
	}
	if strings.Contains(store.putKey, "malicious-name") {
		t.Fatalf("object key contains user-controlled filename: %q", store.putKey)
	}
	if store.putContentType != "image/png" {
		t.Fatalf("content type = %q, want image/png", store.putContentType)
	}
	if !bytes.Equal(store.putBody, image) {
		t.Fatal("stored body differs from uploaded image")
	}
	if !strings.Contains(recorder.Body.String(), "/media/social-network-media/chat/users/7/") {
		t.Fatalf("response does not contain media URL: %s", recorder.Body.String())
	}
}

func TestUploadImageRejectsFilenameDisguisedHTML(t *testing.T) {
	request := multipartRequest(t, "looks-safe.png", []byte("<html><script>alert(1)</script></html>"))
	request = middleware.SetUserIDInContext(request, 7)
	recorder := httptest.NewRecorder()
	store := &fakeMediaStore{}

	NewUploadHandlers(store).UploadImage(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if store.putKey != "" {
		t.Fatalf("invalid upload reached object storage with key %q", store.putKey)
	}
}

func TestDeleteImageEnforcesUserKeyPrefix(t *testing.T) {
	store := &fakeMediaStore{keyFromURL: "chat/users/8/some-image.png"}
	request := httptest.NewRequest(http.MethodDelete, "/upload/delete?path="+url.QueryEscape("/media/social-network-media/chat/users/8/some-image.png"), nil)
	request = middleware.SetUserIDInContext(request, 7)
	recorder := httptest.NewRecorder()

	NewUploadHandlers(store).DeleteImage(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if store.deletedKey != "" {
		t.Fatalf("deleted another user's key %q", store.deletedKey)
	}
}

func TestDeleteImageDeletesOwnedKey(t *testing.T) {
	store := &fakeMediaStore{keyFromURL: "chat/users/7/some-image.png"}
	request := httptest.NewRequest(http.MethodDelete, "/upload/delete?path=owned", nil)
	request = middleware.SetUserIDInContext(request, 7)
	recorder := httptest.NewRecorder()

	NewUploadHandlers(store).DeleteImage(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if store.deletedKey != "chat/users/7/some-image.png" {
		t.Fatalf("deleted key = %q", store.deletedKey)
	}
}

func multipartRequest(t *testing.T, filename string, contents []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/upload/image", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
