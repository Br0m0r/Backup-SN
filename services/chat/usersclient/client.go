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

type Contact struct {
	Profile
	IsMessageRequest bool `json:"is_message_request"`
}

type Directory interface {
	Profiles(context.Context, []int) ([]Profile, error)
	CanStartConversation(context.Context, int, int) (bool, error)
	Contacts(context.Context, int, []int) ([]Contact, error)
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
	profiles := make([]Profile, 0, len(userIDs))
	for start := 0; start < len(userIDs); start += 200 {
		end := min(start+200, len(userIDs))
		values := make([]string, end-start)
		for index, userID := range userIDs[start:end] {
			values[index] = strconv.Itoa(userID)
		}
		var envelope struct {
			Success bool `json:"success"`
			Data    struct {
				Profiles []Profile `json:"profiles"`
			} `json:"data"`
		}
		if err := c.get(ctx, "/internal/v1/users/profiles?ids="+url.QueryEscape(strings.Join(values, ",")), &envelope); err != nil {
			return nil, err
		}
		if !envelope.Success {
			return nil, errors.New("Users profile contract returned an unsuccessful response")
		}
		profiles = append(profiles, envelope.Data.Profiles...)
	}
	return profiles, nil
}

func (c *Client) CanStartConversation(ctx context.Context, senderID, receiverID int) (bool, error) {
	values := url.Values{}
	values.Set("sender_id", strconv.Itoa(senderID))
	values.Set("receiver_id", strconv.Itoa(receiverID))
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			CanStart bool `json:"can_start"`
		} `json:"data"`
	}
	if err := c.get(ctx, "/internal/v1/users/chat/permission?"+values.Encode(), &envelope); err != nil {
		return false, err
	}
	if !envelope.Success {
		return false, errors.New("Users conversation contract returned an unsuccessful response")
	}
	return envelope.Data.CanStart, nil
}

func (c *Client) Contacts(ctx context.Context, userID int, recentSenderIDs []int) ([]Contact, error) {
	values := url.Values{}
	values.Set("user_id", strconv.Itoa(userID))
	if len(recentSenderIDs) > 0 {
		ids := make([]string, len(recentSenderIDs))
		for index, senderID := range recentSenderIDs {
			ids[index] = strconv.Itoa(senderID)
		}
		values.Set("recent_sender_ids", strings.Join(ids, ","))
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Contacts []Contact `json:"contacts"`
		} `json:"data"`
	}
	if err := c.get(ctx, "/internal/v1/users/chat/contacts?"+values.Encode(), &envelope); err != nil {
		return nil, err
	}
	if !envelope.Success {
		return nil, errors.New("Users contacts contract returned an unsuccessful response")
	}
	return envelope.Data.Contacts, nil
}

func (c *Client) get(ctx context.Context, path string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set(serviceauth.HeaderName, c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("call Users contract: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Users contract returned status %d", response.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(target)
}
