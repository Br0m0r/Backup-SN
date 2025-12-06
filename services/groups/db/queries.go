package db

import (
	"database/sql"
	"social-network/services/groups/models"
	"time"
)

// CreateGroup creates a new group
func CreateGroup(db *sql.DB, name string, description, imageURL *string, creatorID int) (*models.Group, error) {
	query := `
		INSERT INTO groups (name, description, image_url, creator_id)
		VALUES (?, ?, ?, ?)
	`
	result, err := db.Exec(query, name, description, imageURL, creatorID)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Add creator as admin member
	memberQuery := `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'admin', 'accepted')
	`
	_, err = db.Exec(memberQuery, id, creatorID)
	if err != nil {
		return nil, err
	}

	return GetGroupByID(db, int(id))
}

// GetGroupByID retrieves a group by ID
func GetGroupByID(db *sql.DB, groupID int) (*models.Group, error) {
	query := `
		SELECT id, name, description, image_url, creator_id, created_at
		FROM groups
		WHERE id = ?
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
				WHERE group_id = g.id AND user_id = ? AND status = 'accepted'
			) THEN 1 ELSE 0 END as is_member,
			CASE WHEN g.creator_id = ? THEN 1 ELSE 0 END as is_creator,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = ? AND status = 'pending'
			) THEN 1 ELSE 0 END as has_pending_request
		FROM groups g
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		WHERE g.id = ?
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
	query := `
		SELECT 
			g.id, g.name, g.description, g.image_url, g.creator_id, g.created_at,
			COUNT(DISTINCT gm.id) as member_count,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = ? AND status = 'accepted'
			) THEN 1 ELSE 0 END as is_member,
			CASE WHEN g.creator_id = ? THEN 1 ELSE 0 END as is_creator,
			CASE WHEN EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = g.id AND user_id = ? AND status = 'pending'
			) THEN 1 ELSE 0 END as has_pending_request
		FROM groups g
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		GROUP BY g.id
		ORDER BY g.created_at DESC
	`

	rows, err := db.Query(query, userID, userID, userID)
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
		SET name = COALESCE(?, name),
			description = COALESCE(?, description),
			image_url = COALESCE(?, image_url)
		WHERE id = ?
	`
	_, err := db.Exec(query, name, description, imageURL, groupID)
	return err
}

// UpdateGroupImage updates only the group's image URL
func UpdateGroupImage(db *sql.DB, groupID int, imageURL string) error {
	query := `UPDATE groups SET image_url = ? WHERE id = ?`
	_, err := db.Exec(query, imageURL, groupID)
	return err
}

// IsGroupCreator checks if user is the group creator
func IsGroupCreator(db *sql.DB, groupID, userID int) (bool, error) {
	query := `SELECT creator_id FROM groups WHERE id = ?`
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
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`
	var count int
	err := db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// InviteMember invites a user to join a group (auto-accepts them)
func InviteMember(db *sql.DB, groupID, userID int) error {
	query := `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'member', 'accepted')
	`
	_, err := db.Exec(query, groupID, userID)
	return err
}

// RequestToJoinGroup creates a pending membership request
func RequestToJoinGroup(db *sql.DB, groupID, userID int) error {
	query := `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'member', 'pending')
	`
	_, err := db.Exec(query, groupID, userID)
	return err
}

// GetPendingRequests gets all pending join requests for a group with user details
func GetPendingRequests(db *sql.DB, groupID int) ([]*models.GroupMember, error) {
	query := `
		SELECT 
			gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.joined_at,
			u.username, u.first_name, u.last_name, u.nickname
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'pending'
		ORDER BY gm.joined_at DESC
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.GroupMember
	for rows.Next() {
		member := &models.GroupMember{}
		var firstName, lastName, nickname *string
		err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Status,
			&member.JoinedAt,
			&member.Username,
			&firstName,
			&lastName,
			&nickname,
		)
		if err != nil {
			return nil, err
		}
		// Assign optional fields
		member.FirstName = firstName
		member.LastName = lastName
		member.Nickname = nickname
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
			WHERE id = ?
		`
		_, err := db.Exec(query, memberID)
		return err
	} else {
		query := `DELETE FROM group_members WHERE id = ?`
		_, err := db.Exec(query, memberID)
		return err
	}
}

// GetGroupMembers retrieves all accepted members of a group with user details
func GetGroupMembers(db *sql.DB, groupID int) ([]*models.GroupMember, error) {
	query := `
		SELECT 
			gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.joined_at,
			u.username, u.first_name, u.last_name, u.nickname
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'accepted'
		ORDER BY gm.joined_at ASC
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.GroupMember
	for rows.Next() {
		member := &models.GroupMember{}
		var firstName, lastName, nickname *string
		err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Status,
			&member.JoinedAt,
			&member.Username,
			&firstName,
			&lastName,
			&nickname,
		)
		if err != nil {
			return nil, err
		}
		// Assign optional fields
		member.FirstName = firstName
		member.LastName = lastName
		member.Nickname = nickname
		members = append(members, member)
	}

	return members, rows.Err()
}

// CreateEvent creates a new event for a group
func CreateEvent(db *sql.DB, groupID, creatorID int, title string, description *string, eventTime time.Time) (*models.Event, error) {
	query := `
		INSERT INTO events (group_id, creator_id, title, description, event_time)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, groupID, creatorID, title, description, eventTime)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetEventByID(db, int(id))
}

// GetEventByID retrieves an event by ID
func GetEventByID(db *sql.DB, eventID int) (*models.Event, error) {
	query := `
		SELECT id, group_id, creator_id, title, description, event_time, created_at
		FROM events
		WHERE id = ?
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
			(SELECT response FROM event_responses WHERE event_id = e.id AND user_id = ?) as user_response
		FROM events e
		LEFT JOIN event_responses er ON e.id = er.event_id
		WHERE e.id = ?
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
			(SELECT response FROM event_responses WHERE event_id = e.id AND user_id = ?) as user_response
		FROM events e
		LEFT JOIN event_responses er ON e.id = er.event_id
		WHERE e.group_id = ?
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
		VALUES (?, ?, ?)
		ON CONFLICT(event_id, user_id) 
		DO UPDATE SET response = ?, created_at = datetime('now')
	`
	_, err := db.Exec(query, eventID, userID, response, response)
	return err
}

// CreateGroupMessage creates a new message in a group chat
func CreateGroupMessage(db *sql.DB, groupID, senderID int, content string) (*models.GroupMessage, error) {
	query := `
		INSERT INTO group_messages (group_id, sender_id, content)
		VALUES (?, ?, ?)
	`
	result, err := db.Exec(query, groupID, senderID, content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetGroupMessageByID(db, int(id))
}

// GetGroupMessageByID retrieves a group message by ID
func GetGroupMessageByID(db *sql.DB, messageID int) (*models.GroupMessage, error) {
	query := `
		SELECT id, group_id, sender_id, content, created_at
		FROM group_messages
		WHERE id = ?
	`
	message := &models.GroupMessage{}
	err := db.QueryRow(query, messageID).Scan(
		&message.ID,
		&message.GroupID,
		&message.SenderID,
		&message.Content,
		&message.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// GetGroupMessages retrieves messages for a group
func GetGroupMessages(db *sql.DB, groupID int, limit int) ([]*models.GroupMessage, error) {
	query := `
		SELECT id, group_id, sender_id, content, created_at
		FROM group_messages
		WHERE group_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.GroupMessage
	for rows.Next() {
		message := &models.GroupMessage{}
		err := rows.Scan(
			&message.ID,
			&message.GroupID,
			&message.SenderID,
			&message.Content,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// GetUsernameByID retrieves username from users table
func GetUsernameByID(db *sql.DB, userID int) (string, error) {
	var username string
	query := `SELECT username FROM users WHERE id = ?`
	err := db.QueryRow(query, userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func RemoveGroupMember(db *sql.DB, groupID, userID int) error {
	query := `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ?
	`
	_, err := db.Exec(query, groupID, userID)
	return err
}
