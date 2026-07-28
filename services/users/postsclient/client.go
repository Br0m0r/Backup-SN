package postsclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"social-network/services/common/serviceauth"
	"social-network/services/users/models"
)

type Reader interface {
	UserPosts(context.Context, int) ([]models.UserPost, error)
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func FromEnvironment(token string) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("POSTS_SERVICE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://post-service:8083"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("POSTS_SERVICE_URL must be an absolute HTTP(S) URL")
	}
	timeout := 3 * time.Second
	if value := strings.TrimSpace(os.Getenv("POSTS_SERVICE_TIMEOUT")); value != "" {
		timeout, err = time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			return nil, errors.New("POSTS_SERVICE_TIMEOUT must be a positive duration")
		}
	}
	return &Client{baseURL: baseURL, token: token, httpClient: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) UserPosts(ctx context.Context, userID int) ([]models.UserPost, error) {
	if userID <= 0 {
		return nil, errors.New("user ID must be positive")
	}
	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.baseURL+"/internal/v1/users/"+strconv.Itoa(userID)+"/posts", nil,
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set(serviceauth.HeaderName, c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call Posts user-post contract: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Posts user-post contract returned status %d", response.StatusCode)
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Posts []models.UserPost `json:"posts"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&envelope); err != nil {
		return nil, err
	}
	if !envelope.Success {
		return nil, errors.New("Posts user-post contract returned an unsuccessful response")
	}
	if envelope.Data.Posts == nil {
		envelope.Data.Posts = []models.UserPost{}
	}
	return envelope.Data.Posts, nil
}
