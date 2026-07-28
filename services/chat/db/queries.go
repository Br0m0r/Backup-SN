package db

import (
	"database/sql"
	"log"
	"social-network/services/chat/models"
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
func GetConversations(messageDB *sql.DB, userID int) ([]models.Conversation, error) {
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

		if err := rows.Scan(
			&conv.UserID,
			&conv.LastMessage,
			&conv.LastMessageAt,
			&conv.UnreadCount,
		); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}
	return conversations, rows.Err()
}

func HasChatHistory(messageDB *sql.DB, senderID, receiverID int) (bool, error) {
	var hasHistory bool
	err := messageDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM messages
			WHERE (sender_id = $1 AND recipient_id = $2)
			   OR (sender_id = $2 AND recipient_id = $1)
		)
	`, senderID, receiverID).Scan(&hasHistory)
	return hasHistory, err
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

func GetRecentSenderIDs(messageDB *sql.DB, userID int) ([]int, error) {
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

func GetContactActivity(messageDB *sql.DB, userID, contactID int) (sql.NullTime, int, error) {
	var lastChatTime sql.NullTime
	var unreadCount int
	err := messageDB.QueryRow(`
		SELECT MAX(created_at),
		       COUNT(*) FILTER (WHERE sender_id = $2 AND recipient_id = $1 AND is_read = FALSE)
		FROM messages
		WHERE (sender_id = $1 AND recipient_id = $2)
		   OR (sender_id = $2 AND recipient_id = $1)
	`, userID, contactID).Scan(&lastChatTime, &unreadCount)
	return lastChatTime, unreadCount, err
}
