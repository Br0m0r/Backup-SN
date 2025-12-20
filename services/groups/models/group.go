package models

import "time"

// Group represents a social network group
type Group struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ImageURL    *string   `json:"image_url,omitempty"`
	CreatorID   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	ID       int       `json:"id"`
	GroupID  int       `json:"group_id"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"`   // "admin" or "member"
	Status   string    `json:"status"` // "pending" or "accepted"
	JoinedAt time.Time `json:"joined_at"`
	// User details (populated when joined with users table)
	Username  string  `json:"username,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
}

// GroupMessage represents a message in a group chat
type GroupMessage struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	SenderID  *int      `json:"sender_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Event represents a group event
type Event struct {
	ID          int       `json:"id"`
	GroupID     int       `json:"group_id"`
	CreatorID   *int      `json:"creator_id,omitempty"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventTime   time.Time `json:"event_time"`
	CreatedAt   time.Time `json:"created_at"`
}

// EventResponse represents a user's RSVP to an event
type EventResponse struct {
	ID        int       `json:"id"`
	EventID   int       `json:"event_id"`
	UserID    int       `json:"user_id"`
	Response  string    `json:"response"` // "going", "not_going", "interested"
	CreatedAt time.Time `json:"created_at"`
}

// CreateGroupRequest represents a request to create a group
type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
}

// InviteMemberRequest represents a request to invite a user to a group
type InviteMemberRequest struct {
	UserID int `json:"user_id"`
}

// JoinGroupRequest represents a user's request to join a group
type JoinGroupRequest struct {
	GroupID int `json:"group_id"`
}

// RespondToRequestRequest represents admin response to join request
type RespondToRequestRequest struct {
	MemberID int  `json:"member_id"`
	Accept   bool `json:"accept"`
}

// GroupInvitation represents a group invitation
type GroupInvitation struct {
	ID               int       `json:"id"`
	GroupID          int       `json:"group_id"`
	UserID           int       `json:"user_id"`
	Role             string    `json:"role,omitempty"`
	Status           string    `json:"status,omitempty"`
	InvitedAt        time.Time `json:"invited_at"`
	GroupName        string    `json:"group_name"`
	GroupDescription *string   `json:"group_description,omitempty"`
	GroupImageURL    *string   `json:"group_image_url,omitempty"`
}

// RespondToInvitationRequest represents user response to group invitation
type RespondToInvitationRequest struct {
	Accept bool `json:"accept"`
}

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	GroupID     int     `json:"group_id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	EventTime   string  `json:"event_time"` // ISO format
}

// EventResponseRequest represents a user's RSVP to an event
type EventResponseRequest struct {
	EventID  int    `json:"event_id"`
	Response string `json:"response"` // "going", "not_going", "interested"
}

// GroupWithDetails includes group info with member count
type GroupWithDetails struct {
	Group
	MemberCount       int  `json:"member_count"`
	IsMember          bool `json:"is_member"`
	IsCreator         bool `json:"is_creator"`
	HasPendingRequest bool `json:"has_pending_request"`
}

// EventWithResponses includes event with response counts
type EventWithResponses struct {
	Event
	CreatorName     string `json:"creator_name,omitempty"`
	GoingCount      int    `json:"going_count"`
	NotGoingCount   int    `json:"not_going_count"`
	InterestedCount int    `json:"interested_count"`
	UserResponse    string `json:"user_response,omitempty"` // Current user's response
}
