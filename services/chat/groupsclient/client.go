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

const defaultTimeout = 3 * time.Second

// Membership is the Groups-owned contract Chat needs to authorize and fan out
// group messages.
type Membership interface {
	IsMember(context.Context, int, int) (bool, error)
	MemberIDs(context.Context, int) ([]int, error)
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
	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("GROUPS_SERVICE_URL must be an absolute HTTP(S) URL")
	}

	timeout := defaultTimeout
	if value := strings.TrimSpace(os.Getenv("GROUPS_SERVICE_TIMEOUT")); value != "" {
		timeout, err = time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			return nil, errors.New("GROUPS_SERVICE_TIMEOUT must be a positive duration")
		}
	}

	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) IsMember(ctx context.Context, groupID, userID int) (bool, error) {
	if groupID <= 0 || userID <= 0 {
		return false, errors.New("group and user IDs must be positive")
	}

	var payload struct {
		IsMember bool `json:"is_member"`
	}
	path := "/internal/v1/groups/" + strconv.Itoa(groupID) + "/members/" + strconv.Itoa(userID)
	if err := c.get(ctx, path, &payload); err != nil {
		return false, err
	}
	return payload.IsMember, nil
}

func (c *Client) MemberIDs(ctx context.Context, groupID int) ([]int, error) {
	if groupID <= 0 {
		return nil, errors.New("group ID must be positive")
	}

	var payload struct {
		MemberIDs []int `json:"member_ids"`
	}
	path := "/internal/v1/groups/" + strconv.Itoa(groupID) + "/members"
	if err := c.get(ctx, path, &payload); err != nil {
		return nil, err
	}
	if payload.MemberIDs == nil {
		payload.MemberIDs = []int{}
	}
	return payload.MemberIDs, nil
}

func (c *Client) get(ctx context.Context, path string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build Groups request: %w", err)
	}
	request.Header.Set(serviceauth.HeaderName, c.token)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("call Groups membership contract: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Groups membership contract returned status %d", response.StatusCode)
	}

	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   string          `json:"error"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode Groups membership response: %w", err)
	}
	if !envelope.Success {
		return errors.New("Groups membership contract returned an unsuccessful response")
	}
	if err := json.Unmarshal(envelope.Data, destination); err != nil {
		return fmt.Errorf("decode Groups membership data: %w", err)
	}
	return nil
}
