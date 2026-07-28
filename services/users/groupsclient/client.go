package groupsclient

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

type Participants interface {
	ParticipantIDs(context.Context, int) ([]int, error)
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func FromEnvironment(token string) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("GROUPS_SERVICE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://group-service:8084"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("GROUPS_SERVICE_URL must be an absolute HTTP(S) URL")
	}
	timeout := 3 * time.Second
	if value := strings.TrimSpace(os.Getenv("GROUPS_SERVICE_TIMEOUT")); value != "" {
		timeout, err = time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			return nil, errors.New("GROUPS_SERVICE_TIMEOUT must be a positive duration")
		}
	}
	return &Client{baseURL: baseURL, token: token, httpClient: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) ParticipantIDs(ctx context.Context, groupID int) ([]int, error) {
	if groupID <= 0 {
		return nil, errors.New("group ID must be positive")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/internal/v1/groups/"+strconv.Itoa(groupID)+"/participants", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set(serviceauth.HeaderName, c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call Groups participant contract: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Groups participant contract returned status %d", response.StatusCode)
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			ParticipantIDs []int `json:"participant_ids"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&envelope); err != nil {
		return nil, err
	}
	if !envelope.Success {
		return nil, errors.New("Groups participant contract returned an unsuccessful response")
	}
	return envelope.Data.ParticipantIDs, nil
}
