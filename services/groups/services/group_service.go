package services

import (
	"database/sql"
	"errors"
	"social-network/services/groups/db"
	"social-network/services/groups/models"
	"time"
)

type GroupService struct {
	database *sql.DB
}

func NewGroupService(database *sql.DB) *GroupService {
	return &GroupService{database: database}
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(req *models.CreateGroupRequest, creatorID int) (*models.Group, error) {
	if req.Name == "" {
		return nil, errors.New("group name is required")
	}

	return db.CreateGroup(s.database, req.Name, req.Description, req.ImageURL, creatorID)
}

// GetGroup retrieves group with details
func (s *GroupService) GetGroup(groupID, userID int) (*models.GroupWithDetails, error) {
	return db.GetGroupWithDetails(s.database, groupID, userID)
}

// GetAllGroups retrieves all groups for browsing
func (s *GroupService) GetAllGroups(userID int) ([]*models.GroupWithDetails, error) {
	return db.GetAllGroups(s.database, userID)
}

// UpdateGroup updates group details (creator only)
func (s *GroupService) UpdateGroup(groupID, userID int, req *models.UpdateGroupRequest) error {
	// Check if user is creator
	isCreator, err := db.IsGroupCreator(s.database, groupID, userID)
	if err != nil {
		return err
	}
	if !isCreator {
		return errors.New("only group creator can update group details")
	}

	return db.UpdateGroup(s.database, groupID, req.Name, req.Description, req.ImageURL)
}

// InviteMember invites a user to join the group (members can invite)
func (s *GroupService) InviteMember(groupID, inviterID, invitedUserID int) error {
	// Check if inviter is a member
	isMember, err := db.IsGroupMember(s.database, groupID, inviterID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("only group members can invite others")
	}

	return db.InviteMember(s.database, groupID, invitedUserID)
}

// RequestToJoin creates a join request (any user can request)
func (s *GroupService) RequestToJoin(groupID, userID int) error {
	return db.RequestToJoinGroup(s.database, groupID, userID)
}

// GetPendingRequests retrieves pending join requests (creator only)
func (s *GroupService) GetPendingRequests(groupID, userID int) ([]*models.GroupMember, error) {
	// Check if user is creator
	isCreator, err := db.IsGroupCreator(s.database, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isCreator {
		return nil, errors.New("only group creator can view pending requests")
	}

	return db.GetPendingRequests(s.database, groupID)
}

// RespondToRequest accepts or rejects a join request (creator only)
func (s *GroupService) RespondToRequest(groupID, memberID, userID int, accept bool) error {
	// Check if user is creator
	isCreator, err := db.IsGroupCreator(s.database, groupID, userID)
	if err != nil {
		return err
	}
	if !isCreator {
		return errors.New("only group creator can respond to requests")
	}

	return db.RespondToJoinRequest(s.database, memberID, accept)
}

// GetGroupMembers retrieves all accepted members
func (s *GroupService) GetGroupMembers(groupID, userID int) ([]*models.GroupMember, error) {
	// Check if user is a member
	isMember, err := db.IsGroupMember(s.database, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("only group members can view member list")
	}

	return db.GetGroupMembers(s.database, groupID)
}

// CreateEvent creates a new event (members can create events)
func (s *GroupService) CreateEvent(req *models.CreateEventRequest, creatorID int) (*models.Event, error) {
	// Check if creator is a member
	isMember, err := db.IsGroupMember(s.database, req.GroupID, creatorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("only group members can create events")
	}

	if req.Title == "" {
		return nil, errors.New("event title is required")
	}

	if req.EventTime == "" {
		return nil, errors.New("event time is required")
	}

	// Parse event time
	eventTime, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		return nil, errors.New("invalid event time format (use ISO 8601/RFC3339)")
	}

	return db.CreateEvent(s.database, req.GroupID, creatorID, req.Title, req.Description, eventTime)
}

// GetEvent retrieves event with response counts
func (s *GroupService) GetEvent(eventID, userID int) (*models.EventWithResponses, error) {
	return db.GetEventWithResponses(s.database, eventID, userID)
}

// GetGroupEvents retrieves all events for a group
func (s *GroupService) GetGroupEvents(groupID, userID int) ([]*models.EventWithResponses, error) {
	// Check if user is a member
	isMember, err := db.IsGroupMember(s.database, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("only group members can view events")
	}

	return db.GetGroupEvents(s.database, groupID, userID)
}

// RespondToEvent creates or updates a user's RSVP to an event
func (s *GroupService) RespondToEvent(req *models.EventResponseRequest, userID int) error {
	// Get event to check group membership
	event, err := db.GetEventByID(s.database, req.EventID)
	if err != nil {
		return err
	}

	// Check if user is a member of the group
	isMember, err := db.IsGroupMember(s.database, event.GroupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("only group members can respond to events")
	}

	// Validate response
	if req.Response != "going" && req.Response != "not_going" && req.Response != "interested" {
		return errors.New("response must be 'going', 'not_going', or 'interested'")
	}

	return db.RespondToEvent(s.database, req.EventID, userID, req.Response)
}

// CreateGroupMessage creates a message in group chat (members only)
func (s *GroupService) CreateGroupMessage(groupID, senderID int, content string) (*models.GroupMessage, error) {
	// Check if sender is a member
	isMember, err := db.IsGroupMember(s.database, groupID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("only group members can send messages")
	}

	if content == "" {
		return nil, errors.New("message content is required")
	}

	return db.CreateGroupMessage(s.database, groupID, senderID, content)
}

// GetGroupMessages retrieves group chat messages (members only)
func (s *GroupService) GetGroupMessages(groupID, userID int, limit int) ([]*models.GroupMessage, error) {
	// Check if user is a member
	isMember, err := db.IsGroupMember(s.database, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("only group members can view messages")
	}

	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	return db.GetGroupMessages(s.database, groupID, limit)
}
