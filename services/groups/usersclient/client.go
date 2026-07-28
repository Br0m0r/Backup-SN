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
	Nickname   *string `json:"nickname"`
}

type Directory interface {
	Profiles(context.Context, []int) ([]Profile, error)
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
		values[index] = strconv.Itoa(userID)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/internal/v1/users/profiles?ids="+url.QueryEscape(strings.Join(values, ",")), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set(serviceauth.HeaderName, c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call Users profile contract: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Users profile contract returned status %d", response.StatusCode)
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Profiles []Profile `json:"profiles"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&envelope); err != nil {
		return nil, err
	}
	if !envelope.Success {
		return nil, errors.New("Users profile contract returned an unsuccessful response")
	}
	return envelope.Data.Profiles, nil
}
