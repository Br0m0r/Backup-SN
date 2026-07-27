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
)

type fakePostMediaStore struct {
	putKey         string
	putBody        []byte
	putContentType string
	deletedKey     string
	keyFromURL     string
}

func (store *fakePostMediaStore) Put(_ context.Context, key string, reader io.Reader, _ int64, contentType string) error {
	store.putKey = key
	store.putContentType = contentType
	store.putBody, _ = io.ReadAll(reader)
	return nil
}

func (store *fakePostMediaStore) Delete(_ context.Context, key string) error {
	store.deletedKey = key
	return nil
}

func (store *fakePostMediaStore) URL(key string) string {
	return "/media/social-network-media/" + key
}

func (store *fakePostMediaStore) KeyFromURL(string) (string, error) {
	return store.keyFromURL, nil
}

func TestPostUploadStoresDetectedImageWithOpaqueName(t *testing.T) {
	image := append([]byte("GIF89a"), bytes.Repeat([]byte{0}, 32)...)
	request := postMultipartRequest(t, "user-controlled.png", image)
	request = withPostUser(request, 12)
	recorder := httptest.NewRecorder()
	store := &fakePostMediaStore{}

	NewUploadHandlers(store).UploadImage(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.HasPrefix(store.putKey, "posts/users/12/") || !strings.HasSuffix(store.putKey, ".gif") {
		t.Fatalf("unexpected object key: %q", store.putKey)
	}
	if strings.Contains(store.putKey, "user-controlled") {
		t.Fatalf("object key contains user filename: %q", store.putKey)
	}
	if store.putContentType != "image/gif" || !bytes.Equal(store.putBody, image) {
		t.Fatalf("stored content type/body did not match detected image")
	}
}

func TestPostUploadRejectsDisguisedText(t *testing.T) {
	request := postMultipartRequest(t, "fake.jpg", []byte("this is plain text"))
	request = withPostUser(request, 12)
	recorder := httptest.NewRecorder()
	store := &fakePostMediaStore{}

	NewUploadHandlers(store).UploadImage(recorder, request)

	if recorder.Code != http.StatusBadRequest || store.putKey != "" {
		t.Fatalf("status = %d, stored key = %q", recorder.Code, store.putKey)
	}
}

func TestPostDeleteRejectsAnotherUsersKey(t *testing.T) {
	store := &fakePostMediaStore{keyFromURL: "posts/users/99/image.png"}
	request := httptest.NewRequest(http.MethodDelete, "/upload/image?path="+url.QueryEscape("/media/social-network-media/posts/users/99/image.png"), nil)
	request = withPostUser(request, 12)
	recorder := httptest.NewRecorder()

	NewUploadHandlers(store).DeleteImage(recorder, request)

	if recorder.Code != http.StatusForbidden || store.deletedKey != "" {
		t.Fatalf("status = %d, deleted key = %q", recorder.Code, store.deletedKey)
	}
}

func TestPostDeleteRemovesOwnedKey(t *testing.T) {
	store := &fakePostMediaStore{keyFromURL: "posts/users/12/image.png"}
	request := httptest.NewRequest(http.MethodDelete, "/upload/image?path=owned", nil)
	request = withPostUser(request, 12)
	recorder := httptest.NewRecorder()

	NewUploadHandlers(store).DeleteImage(recorder, request)

	if recorder.Code != http.StatusOK || store.deletedKey != "posts/users/12/image.png" {
		t.Fatalf("status = %d, deleted key = %q", recorder.Code, store.deletedKey)
	}
}

func postMultipartRequest(t *testing.T, filename string, contents []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatalf("write multipart image: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/upload/image", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func withPostUser(request *http.Request, userID int) *http.Request {
	return request.WithContext(context.WithValue(request.Context(), "userID", userID))
}
