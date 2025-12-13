package services

import (
	"database/sql"
	"errors"
	"social-network/services/common/notify"
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

// UpdateGroupImage updates only the group image (creator only)
func (s *GroupService) UpdateGroupImage(groupID int, imageURL string) error {
	return db.UpdateGroupImage(s.database, groupID, imageURL)
}

// InviteMember invites a user to join the group (members can invite)
func (s *GroupService) InviteMember(groupID, inviterID, invitedUserID int, inviterName string) error {
	// Check if inviter is a member
	isMember, err := db.IsGroupMember(s.database, groupID, inviterID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("only group members can invite others")
	}

	_, err = db.InviteMember(s.database, groupID, invitedUserID)
	if err != nil {
		return err
	}

	// Get group info for notification
	group, err := db.GetGroupWithDetails(s.database, groupID, inviterID)
	if err == nil {
		notify.GroupInvite(invitedUserID, groupID, inviterName, group.Name)
	}

	return nil
}

// RequestToJoin creates a join request (any user can request)
func (s *GroupService) RequestToJoin(groupID, userID int, requesterName string) error {
	err := db.RequestToJoinGroup(s.database, groupID, userID)
	if err != nil {
		return err
	}

	// Get group info and notify creator
	group, err := db.GetGroupByID(s.database, groupID)
	if err == nil {
		notify.GroupJoinRequest(group.CreatorID, groupID, requesterName, group.Name)
	}

	return nil
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

	// Get member info before responding
	members, err := db.GetPendingRequests(s.database, groupID)
	if err != nil {
		return err
	}

	var requesterID int
	for _, m := range members {
		if m.ID == memberID {
			requesterID = m.UserID
			break
		}
	}

	err = db.RespondToJoinRequest(s.database, memberID, accept)
	if err != nil {
		return err
	}

	// Send notification
	group, err := db.GetGroupByID(s.database, groupID)
	if err == nil && requesterID > 0 {
		if accept {
			notify.GroupRequestAccepted(requesterID, groupID, group.Name)
			// Notify creator about new member
			requesterName, err := db.GetUsernameByID(s.database, requesterID)
			if err == nil {
				notify.NewGroupMember(group.CreatorID, groupID, requesterName, group.Name)
			}
		} else {
			notify.GroupRequestRejected(requesterID, groupID, group.Name)
		}
	}

	return nil
}

// GetUserInvitations retrieves all pending invitations for a user
func (s *GroupService) GetUserInvitations(userID int) ([]*models.GroupInvitation, error) {
	return db.GetUserInvitations(s.database, userID)
}

// RespondToInvitation handles user response to group invitation
func (s *GroupService) RespondToInvitation(invitationID, userID int, accept bool) error {
	// Get invitation details BEFORE responding (for notification)
	invitations, err := db.GetUserInvitations(s.database, userID)
	if err != nil {
		return err
	}

	var groupID int
	var groupName string
	var creatorID int
	for _, inv := range invitations {
		if inv.ID == invitationID {
			groupID = inv.GroupID
			groupName = inv.GroupName
			// Get group to find creator
			group, err := db.GetGroupByID(s.database, inv.GroupID)
			if err != nil {
				return err
			}
			creatorID = group.CreatorID
			break
		}
	}

	if groupID == 0 {
		return errors.New("invitation not found")
	}

	// Respond to invitation (db function verifies ownership)
	err = db.RespondToInvitation(s.database, invitationID, userID, accept)
	if err != nil {
		return err
	}

	// Send notification to group creator
	username, err := db.GetUsernameByID(s.database, userID)
	if err == nil {
		if accept {
			notify.GroupInvitationAccepted(creatorID, groupID, username, groupName)
		} else {
			notify.GroupInvitationDeclined(creatorID, groupID, username, groupName)
		}
	}

	return nil
}

// RespondToInvitationByGroupID handles response using groupID from notification
func (s *GroupService) RespondToInvitationByGroupID(groupID, userID int, accept bool) error {
	// Get group details for notification
	group, err := db.GetGroupByID(s.database, groupID)
	if err != nil {
		return err
	}

	// Respond using group ID lookup
	err = db.RespondToInvitationByGroupID(s.database, groupID, userID, accept)
	if err != nil {
		return err
	}

	// Send notification to group creator
	username, err := db.GetUsernameByID(s.database, userID)
	if err == nil {
		if accept {
			notify.GroupInvitationAccepted(group.CreatorID, groupID, username, group.Name)
		} else {
			notify.GroupInvitationDeclined(group.CreatorID, groupID, username, group.Name)
		}
	}

	return nil
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
func (s *GroupService) CreateEvent(req *models.CreateEventRequest, creatorID int, creatorName string) (*models.Event, error) {
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

	event, err := db.CreateEvent(s.database, req.GroupID, creatorID, req.Title, req.Description, eventTime)
	if err != nil {
		return nil, err
	}

	// Get group members and notify them
	members, err := db.GetGroupMembers(s.database, req.GroupID)
	if err == nil {
		group, err := db.GetGroupByID(s.database, req.GroupID)
		if err == nil {
			var memberIDs []int
			for _, member := range members {
				if member.UserID != creatorID { // Don't notify creator
					memberIDs = append(memberIDs, member.UserID)
				}
			}
			if len(memberIDs) > 0 {
				notify.EventCreated(memberIDs, event.ID, creatorName, req.Title, group.Name)
			}
		}
	}

	return event, nil
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
func (s *GroupService) RespondToEvent(req *models.EventResponseRequest, userID int, userName string) error {
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

	err = db.RespondToEvent(s.database, req.EventID, userID, req.Response)
	if err != nil {
		return err
	}

	// Notify event creator
	if event.CreatorID != nil && *event.CreatorID != userID {
		notify.EventResponse(*event.CreatorID, event.ID, userName, event.Title, req.Response)
	}

	return nil
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

func (s *GroupService) LeaveGroup(groupID, userID int) error {
	// Check if user is a member
	isMember, err := db.IsGroupMember(s.database, groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("only group members can leave the group")
	}

	return db.RemoveGroupMember(s.database, groupID, userID)
}
