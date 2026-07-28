package db

import (
	"database/sql"
	"errors"
	"social-network/services/groups/models"
	"strings"
	"time"
)

// CreateGroup creates a new group
func CreateGroup(db *sql.DB, name string, description, imageURL *string, creatorID int) (*models.Group, error) {
	transaction, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()
	var id int
	err = transaction.QueryRow(`
		INSERT INTO groups (name, description, image_url, creator_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, name, description, imageURL, creatorID).Scan(&id)
	if err != nil {
		return nil, err
	}

	// Add creator as admin member
	memberQuery := `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'admin', 'accepted')
	`
	_, err = transaction.Exec(memberQuery, id, creatorID)
	if err != nil {
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	return GetGroupByID(db, id)
}

// GetGroupByID retrieves a group by ID
func GetGroupByID(db *sql.DB, groupID int) (*models.Group, error) {
	query := `
		SELECT id, name, description, image_url, creator_id, created_at
		FROM groups
		WHERE id = $1
	`
	group := &models.Group{}
	err := db.QueryRow(query, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.ImageURL,
		&group.CreatorID,
		&group.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// GetGroupWithDetails retrieves group with member count and user's relationship
func GetGroupWithDetails(db *sql.DB, groupID, userID int) (*models.GroupWithDetails, error) {
	query := `
		SELECT 
			g.id, g.name, g.description, g.image_url, g.creator_id, g.created_at,
			COUNT(DISTINCT gm.id) as member_count,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = $1 AND status = 'accepted'
			) THEN 1 ELSE 0 END as is_member,
			CASE WHEN g.creator_id = $2 THEN 1 ELSE 0 END as is_creator,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = $3 AND status = 'pending'
			) THEN 1 ELSE 0 END as has_pending_request
		FROM groups g
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		WHERE g.id = $4
		GROUP BY g.id
	`

	groupDetails := &models.GroupWithDetails{}
	var isMemberInt, isCreatorInt, hasPendingRequestInt int

	err := db.QueryRow(query, userID, userID, userID, groupID).Scan(
		&groupDetails.ID,
		&groupDetails.Name,
		&groupDetails.Description,
		&groupDetails.ImageURL,
		&groupDetails.CreatorID,
		&groupDetails.CreatedAt,
		&groupDetails.MemberCount,
		&isMemberInt,
		&isCreatorInt,
		&hasPendingRequestInt,
	)
	if err != nil {
		return nil, err
	}

	groupDetails.IsMember = isMemberInt == 1
	groupDetails.IsCreator = isCreatorInt == 1
	groupDetails.HasPendingRequest = hasPendingRequestInt == 1

	return groupDetails, nil
}

// GetAllGroups retrieves all groups (for browsing)
func GetAllGroups(db *sql.DB, userID int) ([]*models.GroupWithDetails, error) {
	return queryGroups(db, userID, "")
}

// SearchGroups retrieves groups whose name or description contains searchQuery.
func SearchGroups(db *sql.DB, userID int, searchQuery string) ([]*models.GroupWithDetails, error) {
	return queryGroups(db, userID, strings.TrimSpace(searchQuery))
}

func queryGroups(db *sql.DB, userID int, searchQuery string) ([]*models.GroupWithDetails, error) {
	query := `
		SELECT 
			g.id, g.name, g.description, g.image_url, g.creator_id, g.created_at,
			COUNT(DISTINCT gm.id) as member_count,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = $1 AND status = 'accepted'
			) THEN 1 ELSE 0 END as is_member,
			CASE WHEN g.creator_id = $2 THEN 1 ELSE 0 END as is_creator,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = $3 AND status = 'pending'
			) THEN 1 ELSE 0 END as has_pending_request
		FROM groups g
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		WHERE ($4 = '' OR lower(g.name) LIKE '%' || lower($4) || '%' OR lower(COALESCE(g.description, '')) LIKE '%' || lower($4) || '%')
		GROUP BY g.id
		ORDER BY g.created_at DESC
	`

	rows, err := db.Query(query, userID, userID, userID, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*models.GroupWithDetails
	for rows.Next() {
		groupDetails := &models.GroupWithDetails{}
		var isMemberInt, isCreatorInt, hasPendingRequestInt int

		err := rows.Scan(
			&groupDetails.ID,
			&groupDetails.Name,
			&groupDetails.Description,
			&groupDetails.ImageURL,
			&groupDetails.CreatorID,
			&groupDetails.CreatedAt,
			&groupDetails.MemberCount,
			&isMemberInt,
			&isCreatorInt,
			&hasPendingRequestInt,
		)
		if err != nil {
			return nil, err
		}

		groupDetails.IsMember = isMemberInt == 1
		groupDetails.IsCreator = isCreatorInt == 1
		groupDetails.HasPendingRequest = hasPendingRequestInt == 1

		groups = append(groups, groupDetails)
	}

	return groups, rows.Err()
}

// UpdateGroup updates group details (only creator can do this)
func UpdateGroup(db *sql.DB, groupID int, name, description, imageURL *string) error {
	query := `
		UPDATE groups
		SET name = COALESCE($1, name),
			description = COALESCE($2, description),
			image_url = COALESCE($3, image_url)
		WHERE id = $4
	`
	_, err := db.Exec(query, name, description, imageURL, groupID)
	return err
}

// UpdateGroupImage updates only the group's image URL
func UpdateGroupImage(db *sql.DB, groupID int, imageURL string) error {
	query := `UPDATE groups SET image_url = $1 WHERE id = $2`
	_, err := db.Exec(query, imageURL, groupID)
	return err
}

// IsGroupCreator checks if user is the group creator
func IsGroupCreator(db *sql.DB, groupID, userID int) (bool, error) {
	query := `SELECT creator_id FROM groups WHERE id = $1`
	var creatorID int
	err := db.QueryRow(query, groupID).Scan(&creatorID)
	if err != nil {
		return false, err
	}
	return creatorID == userID, nil
}

// IsGroupMember checks if user is an accepted member of the group
func IsGroupMember(db *sql.DB, groupID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM group_members
		WHERE group_id = $1 AND user_id = $2 AND status = 'accepted'
	`
	var count int
	err := db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAcceptedGroupMemberIDs returns the identities authorized to participate
// in a group's service-to-service workflows.
func GetAcceptedGroupMemberIDs(db *sql.DB, groupID int) ([]int, error) {
	rows, err := db.Query(`
		SELECT user_id
		FROM group_members
		WHERE group_id = $1 AND status = 'accepted'
		ORDER BY user_id
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberIDs := make([]int, 0)
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		memberIDs = append(memberIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memberIDs, nil
}

// GetGroupParticipantIDs returns every identity with an existing membership
// record, including pending and invited users.
func GetGroupParticipantIDs(db *sql.DB, groupID int) ([]int, error) {
	rows, err := db.Query(`
		SELECT user_id FROM group_members
		WHERE group_id = $1
		ORDER BY user_id
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userIDs := make([]int, 0)
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, rows.Err()
}

// InviteMember invites a user to join a group (auto-accepts them)
func InviteMember(db *sql.DB, groupID, userID int) (int, error) {
	query := `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'member', 'invited')
		RETURNING id
	`
	var invitationID int
	err := db.QueryRow(query, groupID, userID).Scan(&invitationID)
	return invitationID, err
}

// RequestToJoinGroup creates a pending membership request
func RequestToJoinGroup(db *sql.DB, groupID, userID int) error {
	query := `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, 'member', 'pending')
	`
	_, err := db.Exec(query, groupID, userID)
	return err
}

// GetPendingRequests gets all pending join requests for a group with user details
func GetPendingRequests(db *sql.DB, groupID int) ([]*models.GroupMember, error) {
	query := `
		SELECT id, group_id, user_id, role, status, joined_at
		FROM group_members
		WHERE group_id = $1 AND status = 'pending'
		ORDER BY joined_at DESC
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.GroupMember
	for rows.Next() {
		member := &models.GroupMember{}
		err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Status,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// RespondToJoinRequest accepts or rejects a pending membership
func RespondToJoinRequest(db *sql.DB, memberID int, accept bool) error {
	if accept {
		query := `
			UPDATE group_members
			SET status = 'accepted'
			WHERE id = $1
		`
		_, err := db.Exec(query, memberID)
		return err
	} else {
		query := `DELETE FROM group_members WHERE id = $1`
		_, err := db.Exec(query, memberID)
		return err
	}
}

// GetUserInvitations retrieves all pending invitations for a user
func GetUserInvitations(db *sql.DB, userID int) ([]*models.GroupInvitation, error) {
	query := `
		SELECT 
			gm.id, gm.group_id, gm.user_id, gm.joined_at,
			g.name, g.description, g.image_url
		FROM group_members gm
		JOIN groups g ON gm.group_id = g.id
		WHERE gm.user_id = $1 AND gm.status = 'invited'
		ORDER BY gm.joined_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*models.GroupInvitation
	for rows.Next() {
		inv := &models.GroupInvitation{}
		var description, imageURL *string
		err := rows.Scan(
			&inv.ID,
			&inv.GroupID,
			&inv.UserID,
			&inv.InvitedAt,
			&inv.GroupName,
			&description,
			&imageURL,
		)
		if err != nil {
			return nil, err
		}
		inv.GroupDescription = description
		inv.GroupImageURL = imageURL
		invitations = append(invitations, inv)
	}

	return invitations, rows.Err()
}

// RespondToInvitation accepts or rejects an invitation
func RespondToInvitation(db *sql.DB, invitationID, userID int, accept bool) error {
	// First verify the invitation belongs to this user
	var count int
	checkQuery := `SELECT COUNT(*) FROM group_members WHERE id = $1 AND user_id = $2 AND status = 'invited'`
	err := db.QueryRow(checkQuery, invitationID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("invitation not found or already responded")
	}

	if accept {
		query := `
			UPDATE group_members
			SET status = 'accepted'
			WHERE id = $1
		`
		_, err := db.Exec(query, invitationID)
		return err
	} else {
		query := `DELETE FROM group_members WHERE id = $1`
		_, err := db.Exec(query, invitationID)
		return err
	}
}

// RespondToInvitationByGroupID responds to invitation using group_id instead of invitation_id
func RespondToInvitationByGroupID(db *sql.DB, groupID, userID int, accept bool) error {
	// Find the invitation ID first
	var invitationID int
	query := `SELECT id FROM group_members WHERE group_id = $1 AND user_id = $2 AND status = 'invited'`
	err := db.QueryRow(query, groupID, userID).Scan(&invitationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invitation not found or already responded")
		}
		return err
	}

	// Use the existing function
	return RespondToInvitation(db, invitationID, userID, accept)
}

// GetGroupMembers retrieves all accepted members of a group with user details
func GetGroupMembers(db *sql.DB, groupID int) ([]*models.GroupMember, error) {
	query := `
		SELECT id, group_id, user_id, role, status, joined_at
		FROM group_members
		WHERE group_id = $1 AND status = 'accepted'
		ORDER BY joined_at ASC
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.GroupMember
	for rows.Next() {
		member := &models.GroupMember{}
		err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Status,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// CreateEvent creates a new event for a group
func CreateEvent(db *sql.DB, groupID, creatorID int, title string, description *string, eventTime time.Time) (*models.Event, error) {
	query := `
		INSERT INTO events (group_id, creator_id, title, description, event_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int
	err := db.QueryRow(query, groupID, creatorID, title, description, eventTime).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetEventByID(db, id)
}

// GetEventByID retrieves an event by ID
func GetEventByID(db *sql.DB, eventID int) (*models.Event, error) {
	query := `
		SELECT id, group_id, creator_id, title, description, event_time, created_at
		FROM events
		WHERE id = $1
	`
	event := &models.Event{}
	err := db.QueryRow(query, eventID).Scan(
		&event.ID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&event.EventTime,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// GetEventWithResponses retrieves event with response counts and user's response
func GetEventWithResponses(db *sql.DB, eventID, userID int) (*models.EventWithResponses, error) {
	query := `
		SELECT 
			e.id, e.group_id, e.creator_id, e.title, e.description, e.event_time, e.created_at,
			COUNT(CASE WHEN er.response = 'going' THEN 1 END) as going_count,
			COUNT(CASE WHEN er.response = 'not_going' THEN 1 END) as not_going_count,
			COUNT(CASE WHEN er.response = 'interested' THEN 1 END) as interested_count,
			(SELECT response FROM event_responses WHERE event_id = e.id AND user_id = $1) as user_response
		FROM events e
		LEFT JOIN event_responses er ON e.id = er.event_id
		WHERE e.id = $2
		GROUP BY e.id
	`

	eventWithResponses := &models.EventWithResponses{}
	var userResponse sql.NullString

	err := db.QueryRow(query, userID, eventID).Scan(
		&eventWithResponses.ID,
		&eventWithResponses.GroupID,
		&eventWithResponses.CreatorID,
		&eventWithResponses.Title,
		&eventWithResponses.Description,
		&eventWithResponses.EventTime,
		&eventWithResponses.CreatedAt,
		&eventWithResponses.GoingCount,
		&eventWithResponses.NotGoingCount,
		&eventWithResponses.InterestedCount,
		&userResponse,
	)
	if err != nil {
		return nil, err
	}

	if userResponse.Valid {
		eventWithResponses.UserResponse = userResponse.String
	}

	return eventWithResponses, nil
}

// GetGroupEvents retrieves all events for a group
func GetGroupEvents(db *sql.DB, groupID, userID int) ([]*models.EventWithResponses, error) {
	query := `
		SELECT 
			e.id, e.group_id, e.creator_id, e.title, e.description, e.event_time, e.created_at,
			COUNT(CASE WHEN er.response = 'going' THEN 1 END) as going_count,
			COUNT(CASE WHEN er.response = 'not_going' THEN 1 END) as not_going_count,
			COUNT(CASE WHEN er.response = 'interested' THEN 1 END) as interested_count,
			(SELECT response FROM event_responses WHERE event_id = e.id AND user_id = $1) as user_response
		FROM events e
		LEFT JOIN event_responses er ON e.id = er.event_id
		WHERE e.group_id = $2
		GROUP BY e.id
		ORDER BY e.event_time ASC
	`

	rows, err := db.Query(query, userID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.EventWithResponses
	for rows.Next() {
		event := &models.EventWithResponses{}
		var userResponse sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.CreatedAt,
			&event.GoingCount,
			&event.NotGoingCount,
			&event.InterestedCount,
			&userResponse,
		)
		if err != nil {
			return nil, err
		}

		if userResponse.Valid {
			event.UserResponse = userResponse.String
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// RespondToEvent creates or updates a user's response to an event
func RespondToEvent(db *sql.DB, eventID, userID int, response string) error {
	query := `
		INSERT INTO event_responses (event_id, user_id, response)
		VALUES ($1, $2, $3)
		ON CONFLICT(event_id, user_id) 
		DO UPDATE SET response = EXCLUDED.response, created_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, eventID, userID, response)
	return err
}

func RemoveGroupMember(db *sql.DB, groupID, userID int) error {
	query := `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`
	_, err := db.Exec(query, groupID, userID)
	return err
}
