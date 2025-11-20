package db

import (
	"database/sql"
	"log"
	"social-network/services/chat/models"
)

// SaveMessage stores a new message in the database
func SaveMessage(db *sql.DB, msg *models.Message) error {
	query := `
		INSERT INTO messages (sender_id, recipient_id, content, is_read, created_at, image_path)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.IsRead, msg.CreatedAt, msg.ImagePath)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = int(id)
	return nil
}

// GetChatHistory retrieves all messages between two users
func GetChatHistory(db *sql.DB, user1ID, user2ID int, limit int) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, content, is_read, created_at, image_path
		FROM messages
		WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, user1ID, user2ID, user2ID, user1ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.IsRead, &msg.CreatedAt, &msg.ImagePath)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// MarkAsRead marks all messages from a specific sender to receiver as read
func MarkAsRead(db *sql.DB, senderID, receiverID int) error {
	query := `
		UPDATE messages 
		SET is_read = 1 
		WHERE sender_id = ? AND recipient_id = ? AND is_read = 0
	`
	_, err := db.Exec(query, senderID, receiverID)
	return err
}

// GetConversations retrieves all conversations for a user with last message and unread count
func GetConversations(db *sql.DB, userID int) ([]models.Conversation, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.first_name,
			u.last_name,
			u.nickname,
			m.content as last_message,
			m.created_at as last_message_at,
			(SELECT COUNT(*) 
			 FROM messages 
			 WHERE sender_id = u.id 
			   AND recipient_id = ? 
			   AND is_read = 0) as unread_count
		FROM (
			SELECT DISTINCT 
				CASE 
					WHEN sender_id = ? THEN recipient_id 
					ELSE sender_id 
				END as other_user_id
			FROM messages
			WHERE sender_id = ? OR recipient_id = ?
		) conv
		JOIN users u ON u.id = conv.other_user_id
		LEFT JOIN messages m ON m.id = (
			SELECT id FROM messages
			WHERE (sender_id = ? AND recipient_id = u.id)
			   OR (sender_id = u.id AND recipient_id = ?)
			ORDER BY created_at DESC
			LIMIT 1
		)
		ORDER BY m.created_at DESC
	`

	rows, err := db.Query(query, userID, userID, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		var firstName, lastName, nickname sql.NullString

		err := rows.Scan(
			&conv.UserID,
			&conv.Username,
			&firstName,
			&lastName,
			&nickname,
			&conv.LastMessage,
			&conv.LastMessageAt,
			&conv.UnreadCount,
		)
		if err != nil {
			log.Printf("Error scanning conversation: %v", err)
			continue
		}

		if firstName.Valid {
			conv.FirstName = &firstName.String
		}
		if lastName.Valid {
			conv.LastName = &lastName.String
		}
		if nickname.Valid {
			conv.Nickname = &nickname.String
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// CanChat checks if a user can chat with another user
// Rules:
// - If receiver has public profile: sender must be following receiver (one-way)
// - If receiver has private profile: BOTH must be following each other (mutual)
func CanChat(db *sql.DB, senderID, receiverID int) (bool, error) {
	query := `
		SELECT 
			CASE 
				-- If receiver has public profile, sender just needs to follow them
				WHEN (SELECT is_public_profile FROM users WHERE id = ?) = 1 
					AND EXISTS (
						SELECT 1 FROM follows 
						WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
					) THEN 1
				-- If receiver has private profile, need mutual follows
				WHEN (SELECT is_public_profile FROM users WHERE id = ?) = 0 
					AND EXISTS (
						SELECT 1 FROM follows 
						WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
					)
					AND EXISTS (
						SELECT 1 FROM follows 
						WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
					) THEN 1
				ELSE 0
			END as can_chat
	`

	var canChat int
	err := db.QueryRow(query,
		receiverID, senderID, receiverID, // public profile check
		receiverID, senderID, receiverID, receiverID, senderID). // private profile check
		Scan(&canChat)
	if err != nil {
		return false, err
	}

	return canChat == 1, nil
}

// GetUnreadCount returns the total number of unread messages for a user
func GetUnreadCount(db *sql.DB, userID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM messages 
		WHERE recipient_id = ? AND is_read = 0
	`

	var count int
	err := db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// SaveGroupMessage stores a new group message in the database
func SaveGroupMessage(db *sql.DB, msg *models.GroupMessage) error {
	query := `
		INSERT INTO group_messages (group_id, sender_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`
	result, err := db.Exec(query, msg.GroupID, msg.SenderID, msg.Content, msg.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = int(id)
	return nil
}

// GetGroupChatHistory retrieves messages from a group
func GetGroupChatHistory(db *sql.DB, groupID int, limit int) ([]models.GroupMessage, error) {
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

	var messages []models.GroupMessage
	for rows.Next() {
		var msg models.GroupMessage
		err := rows.Scan(&msg.ID, &msg.GroupID, &msg.SenderID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning group message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// IsGroupMember checks if a user is an accepted member of a group
func IsGroupMember(db *sql.DB, groupID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM group_members 
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`

	var count int
	err := db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetGroupMembers retrieves all accepted member IDs for a group
func GetGroupMembers(db *sql.DB, groupID int) ([]int, error) {
	query := `
		SELECT user_id 
		FROM group_members 
		WHERE group_id = ? AND status = 'accepted'
	`

	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Error scanning group member: %v", err)
			continue
		}
		members = append(members, userID)
	}

	return members, nil
}

// GetAvailableContacts retrieves all users the current user can chat with
// Returns users ordered by: 1) Recent chat activity, 2) Alphabetically
// Rules:
// - Public profiles: Show if current user is following them
// - Private profiles: Show only if BOTH users follow each other (mutual)
func GetAvailableContacts(db *sql.DB, userID int) ([]models.ChatContact, error) {
	query := `
        SELECT DISTINCT
            u.id,
            u.username,
            u.first_name,
            u.last_name,
            u.nickname,
            u.avatar_path,
            -- Get last chat time with this user
            (SELECT MAX(created_at) 
             FROM messages 
             WHERE (sender_id = ? AND recipient_id = u.id) 
                OR (sender_id = u.id AND recipient_id = ?)
            ) as last_chat_time,
            -- Count unread messages from this user
            (SELECT COUNT(*) 
             FROM messages 
             WHERE sender_id = u.id AND recipient_id = ? AND is_read = 0
            ) as unread_count
        FROM users u
        -- Current user must be following this contact
        INNER JOIN follows f1 ON f1.following_id = u.id 
            AND f1.follower_id = ? 
            AND f1.status = 'accepted'
        WHERE u.id != ?
            -- Filter: if contact has private profile, they must also follow back
            AND (
                u.is_public_profile = 1
                OR EXISTS (
                    SELECT 1 FROM follows f2 
                    WHERE f2.follower_id = u.id 
                        AND f2.following_id = ? 
                        AND f2.status = 'accepted'
                )
            )
        ORDER BY 
            -- Users with chat history first
            CASE WHEN last_chat_time IS NOT NULL THEN 0 ELSE 1 END,
            -- Most recent chats first
            last_chat_time DESC,
            -- Alphabetically
            u.username ASC
    `

	rows, err := db.Query(query,
		userID, userID, // last_chat_time subquery
		userID, // unread_count subquery
		userID, // f1 join (user follows contact)
		userID, // WHERE clause (exclude self)
		userID) // WHERE EXISTS (mutual follow for private profiles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.ChatContact
	for rows.Next() {
		var contact models.ChatContact
		var firstName, lastName, nickname, avatar, lastChatTime sql.NullString

		err := rows.Scan(
			&contact.UserID,
			&contact.Username,
			&firstName,
			&lastName,
			&nickname,
			&avatar,
			&lastChatTime,
			&contact.UnreadCount,
		)
		if err != nil {
			log.Printf("Error scanning contact: %v", err)
			continue
		}

		// Handle nullable fields
		if firstName.Valid {
			contact.FirstName = firstName.String
		}
		if lastName.Valid {
			contact.LastName = lastName.String
		}
		if nickname.Valid {
			contact.Nickname = nickname.String
		}
		if avatar.Valid {
			contact.Avatar = avatar.String
		}
		if lastChatTime.Valid && lastChatTime.String != "" {
			contact.HasChatHistory = true
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}
