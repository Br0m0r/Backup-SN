package usersclient

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
)

type Profile struct {
	ID         int     `json:"id"`
	Username   string  `json:"username"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	AvatarPath *string `json:"avatar_path"`
}

type Directory interface {
	Profiles(context.Context, []int) ([]Profile, error)
	SearchProfiles(context.Context, string) ([]Profile, error)
	FollowingIDs(context.Context, int) ([]int, error)
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func FromEnvironment(token string) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("USERS_SERVICE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://user-service:8082"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("USERS_SERVICE_URL must be an absolute HTTP(S) URL")
	}
	timeout := 3 * time.Second
	if value := strings.TrimSpace(os.Getenv("USERS_SERVICE_TIMEOUT")); value != "" {
		timeout, err = time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			return nil, errors.New("USERS_SERVICE_TIMEOUT must be a positive duration")
		}
	}
	return &Client{baseURL: baseURL, token: token, httpClient: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) Profiles(ctx context.Context, userIDs []int) ([]Profile, error) {
	if len(userIDs) == 0 {
		return []Profile{}, nil
	}
	values := make([]string, len(userIDs))
	for index, userID := range userIDs {
		if userID <= 0 {
			return nil, errors.New("user IDs must be positive")
		}
		values[index] = strconv.Itoa(userID)
	}
	var payload struct {
		Profiles []Profile `json:"profiles"`
	}
	path := "/internal/v1/users/profiles?ids=" + url.QueryEscape(strings.Join(values, ","))
	if err := c.get(ctx, path, &payload); err != nil {
		return nil, err
	}
	return payload.Profiles, nil
}

func (c *Client) SearchProfiles(ctx context.Context, query string) ([]Profile, error) {
	var payload struct {
		Profiles []Profile `json:"profiles"`
	}
	path := "/internal/v1/users/profiles?q=" + url.QueryEscape(strings.TrimSpace(query))
	if err := c.get(ctx, path, &payload); err != nil {
		return nil, err
	}
	return payload.Profiles, nil
}

func (c *Client) FollowingIDs(ctx context.Context, userID int) ([]int, error) {
	if userID <= 0 {
		return nil, errors.New("user ID must be positive")
	}
	var payload struct {
		FollowingIDs []int `json:"following_ids"`
	}
	path := "/internal/v1/users/" + strconv.Itoa(userID) + "/following"
	if err := c.get(ctx, path, &payload); err != nil {
		return nil, err
	}
	return payload.FollowingIDs, nil
}

func (c *Client) get(ctx context.Context, path string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set(serviceauth.HeaderName, c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("call Users read contract: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Users read contract returned status %d", response.StatusCode)
	}
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&envelope); err != nil {
		return err
	}
	if !envelope.Success {
		return errors.New("Users read contract returned an unsuccessful response")
	}
	return json.Unmarshal(envelope.Data, destination)
}
