package db

import (
	"database/sql"
	"log"
	"social-network/services/chat/models"
	"sort"
	"strings"
	"time"
)

// SaveMessage stores a new message in the database
func SaveMessage(db *sql.DB, msg *models.Message) error {
	query := `
		INSERT INTO messages (sender_id, recipient_id, content, is_read, created_at, image_path)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return db.QueryRow(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.IsRead, msg.CreatedAt, msg.ImagePath).Scan(&msg.ID)
}

// GetChatHistory retrieves all messages between two users
func GetChatHistory(db *sql.DB, user1ID, user2ID int, limit int) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, content, is_read, created_at, image_path
		FROM messages
		WHERE (sender_id = $1 AND recipient_id = $2) OR (sender_id = $2 AND recipient_id = $1)
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := db.Query(query, user1ID, user2ID, limit)
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
		SET is_read = TRUE
		WHERE sender_id = $1 AND recipient_id = $2 AND is_read = FALSE
	`
	_, err := db.Exec(query, senderID, receiverID)
	return err
}

// GetConversations retrieves all conversations for a user with last message and unread count
func GetConversations(messageDB, identityDB *sql.DB, userID int) ([]models.Conversation, error) {
	rows, err := messageDB.Query(`
		SELECT
			other_user_id,
			(SELECT content FROM messages
			 WHERE (sender_id = $1 AND recipient_id = other_user_id)
			    OR (sender_id = other_user_id AND recipient_id = $1)
			 ORDER BY created_at DESC, id DESC LIMIT 1),
			(SELECT created_at FROM messages
			 WHERE (sender_id = $1 AND recipient_id = other_user_id)
			    OR (sender_id = other_user_id AND recipient_id = $1)
			 ORDER BY created_at DESC, id DESC LIMIT 1),
			(SELECT COUNT(*) FROM messages
			 WHERE sender_id = other_user_id AND recipient_id = $1 AND is_read = FALSE)
		FROM (
			SELECT DISTINCT CASE WHEN sender_id = $1 THEN recipient_id ELSE sender_id END AS other_user_id
			FROM messages
			WHERE sender_id = $1 OR recipient_id = $1
		) conversations
		ORDER BY 3 DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]models.Conversation, 0)
	for rows.Next() {
		var conv models.Conversation
		var firstName, lastName, nickname sql.NullString

		if err := rows.Scan(
			&conv.UserID,
			&conv.LastMessage,
			&conv.LastMessageAt,
			&conv.UnreadCount,
		); err != nil {
			return nil, err
		}
		err := identityDB.QueryRow(`
			SELECT username, first_name, last_name, nickname
			FROM users WHERE id = ?
		`, conv.UserID).Scan(&conv.Username, &firstName, &lastName, &nickname)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
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
	return conversations, rows.Err()
}

// CanChat checks if a user can chat with another user
// Rules:
// - If there's existing message history: allow reply (regardless of follow status)
// - If receiver has public profile: sender must be following receiver (one-way)
// - If receiver has private profile: BOTH must be following each other (mutual)
func CanChat(messageDB, identityDB *sql.DB, senderID, receiverID int) (bool, error) {
	var hasHistory bool
	if err := messageDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM messages
			WHERE (sender_id = $1 AND recipient_id = $2)
			   OR (sender_id = $2 AND recipient_id = $1)
		)
	`, senderID, receiverID).Scan(&hasHistory); err != nil {
		return false, err
	}
	if hasHistory {
		return true, nil
	}

	query := `
		SELECT 
			CASE 
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
	err := identityDB.QueryRow(query,
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
		WHERE recipient_id = $1 AND is_read = FALSE
	`

	var count int
	err := db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// SaveGroupMessage stores a new group message in the database
func SaveGroupMessage(db *sql.DB, msg *models.GroupMessage) error {
	query := `
		INSERT INTO group_messages (group_id, sender_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return db.QueryRow(query, msg.GroupID, msg.SenderID, msg.Content, msg.CreatedAt).Scan(&msg.ID)
}

// GetGroupChatHistory retrieves messages from a group
func GetGroupChatHistory(db *sql.DB, groupID int, limit int) ([]models.GroupMessage, error) {
	query := `
		SELECT id, group_id, sender_id, content, created_at
		FROM group_messages
		WHERE group_id = $1
		ORDER BY created_at DESC
		LIMIT $2
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

// GetAvailableContacts retrieves all users the current user can chat with
// Returns users ordered by: 1) Recent chat activity, 2) Alphabetically
// Rules:
// - Users who have messaged you (in last 30 days)
// - Public profiles: Show if current user is following them
// - Private profiles: Show only if BOTH users follow each other (mutual)
func GetAvailableContacts(messageDB, identityDB *sql.DB, userID int) ([]models.ChatContact, error) {
	recentSenderIDs, err := getRecentSenderIDs(messageDB, userID)
	if err != nil {
		return nil, err
	}
	recentClause := ""
	arguments := []any{userID, userID, userID, userID}
	if len(recentSenderIDs) > 0 {
		placeholders := make([]string, len(recentSenderIDs))
		for i, senderID := range recentSenderIDs {
			placeholders[i] = "?"
			arguments = append(arguments, senderID)
		}
		recentClause = " OR u.id IN (" + strings.Join(placeholders, ",") + ")"
	}

	query := `
		SELECT DISTINCT
			u.id, u.username, u.first_name, u.last_name, u.nickname, u.avatar_path,
			CASE WHEN NOT EXISTS (
				SELECT 1 FROM follows
				WHERE follower_id = ? AND following_id = u.id AND status = 'accepted'
			) THEN 1 ELSE 0 END
		FROM users u
		WHERE u.id != ?
		  AND (
			(
				EXISTS (
					SELECT 1 FROM follows f1
					WHERE f1.following_id = u.id AND f1.follower_id = ? AND f1.status = 'accepted'
				)
				AND (
					u.is_public_profile = 1
					OR EXISTS (
						SELECT 1 FROM follows f2
						WHERE f2.follower_id = u.id AND f2.following_id = ? AND f2.status = 'accepted'
					)
				)
			)
			%s
		  )
	`
	rows, err := identityDB.Query(strings.Replace(query, "%s", recentClause, 1), arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type contactWithTime struct {
		contact      models.ChatContact
		lastChatTime sql.NullTime
	}
	results := make([]contactWithTime, 0)
	for rows.Next() {
		var contact models.ChatContact
		var firstName, lastName, nickname, avatar sql.NullString
		var isMessageRequest int

		if err := rows.Scan(
			&contact.UserID,
			&contact.Username,
			&firstName,
			&lastName,
			&nickname,
			&avatar,
			&isMessageRequest,
		); err != nil {
			return nil, err
		}
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
			contact.AvatarPath = avatar.String
		}
		var lastChatTime sql.NullTime
		if err := messageDB.QueryRow(`
			SELECT MAX(created_at),
			       COUNT(*) FILTER (WHERE sender_id = $2 AND recipient_id = $1 AND is_read = FALSE)
			FROM messages
			WHERE (sender_id = $1 AND recipient_id = $2)
			   OR (sender_id = $2 AND recipient_id = $1)
		`, userID, contact.UserID).Scan(&lastChatTime, &contact.UnreadCount); err != nil {
			return nil, err
		}
		contact.HasChatHistory = lastChatTime.Valid
		contact.IsMessageRequest = isMessageRequest == 1
		results = append(results, contactWithTime{contact: contact, lastChatTime: lastChatTime})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(results, func(i, j int) bool {
		left, right := results[i], results[j]
		if left.lastChatTime.Valid != right.lastChatTime.Valid {
			return left.lastChatTime.Valid
		}
		if left.lastChatTime.Valid && !left.lastChatTime.Time.Equal(right.lastChatTime.Time) {
			return left.lastChatTime.Time.After(right.lastChatTime.Time)
		}
		return strings.ToLower(left.contact.Username) < strings.ToLower(right.contact.Username)
	})
	contacts := make([]models.ChatContact, 0, len(results))
	for _, result := range results {
		contacts = append(contacts, result.contact)
	}
	return contacts, nil
}

func getRecentSenderIDs(messageDB *sql.DB, userID int) ([]int, error) {
	rows, err := messageDB.Query(`
		SELECT DISTINCT sender_id
		FROM messages
		WHERE recipient_id = $1 AND created_at >= $2
	`, userID, time.Now().Add(-30*24*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	senderIDs := make([]int, 0)
	for rows.Next() {
		var senderID int
		if err := rows.Scan(&senderID); err != nil {
			return nil, err
		}
		senderIDs = append(senderIDs, senderID)
	}
	return senderIDs, rows.Err()
}
