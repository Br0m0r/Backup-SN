package db

import (
	"database/sql"
	"social-network/services/notifications/models"
)

// CreateNotification inserts a new notification into the database
func CreateNotification(database *sql.DB, notif *models.CreateNotificationRequest) (*models.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, type, related_id, content)
		VALUES (?, ?, ?, ?)
	`

	result, err := database.Exec(query, notif.UserID, notif.Type, notif.RelatedID, notif.Content)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get the created notification
	return GetNotificationByID(database, int(id))
}

// GetNotificationByID retrieves a notification by ID
func GetNotificationByID(database *sql.DB, id int) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, related_id, content, is_read, created_at
		FROM notifications
		WHERE id = ?
	`

	var notif models.Notification
	err := database.QueryRow(query, id).Scan(
		&notif.ID,
		&notif.UserID,
		&notif.Type,
		&notif.RelatedID,
		&notif.Content,
		&notif.IsRead,
		&notif.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &notif, nil
}

// GetUserNotifications retrieves all notifications for a user
func GetUserNotifications(database *sql.DB, userID int, limit, offset int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, related_id, content, is_read, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []models.Notification{}
	for rows.Next() {
		var notif models.Notification
		err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&notif.Type,
			&notif.RelatedID,
			&notif.Content,
			&notif.IsRead,
			&notif.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

// GetUnreadNotifications retrieves unread notifications for a user
func GetUnreadNotifications(database *sql.DB, userID int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, related_id, content, is_read, created_at
		FROM notifications
		WHERE user_id = ? AND is_read = 0
		ORDER BY created_at DESC
	`

	rows, err := database.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []models.Notification{}
	for rows.Next() {
		var notif models.Notification
		err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&notif.Type,
			&notif.RelatedID,
			&notif.Content,
			&notif.IsRead,
			&notif.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func GetUnreadCount(database *sql.DB, userID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = ? AND is_read = 0
	`

	var count int
	err := database.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func MarkAsRead(database *sql.DB, notificationID, userID int) error {
	query := `
		UPDATE notifications
		SET is_read = 1
		WHERE id = ? AND user_id = ?
	`

	_, err := database.Exec(query, notificationID, userID)
	return err
}

// MarkAllAsRead marks all notifications as read for a user
func MarkAllAsRead(database *sql.DB, userID int) error {
	query := `
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ? AND is_read = 0
	`

	_, err := database.Exec(query, userID)
	return err
}

// DeleteNotification deletes a notification
func DeleteNotification(database *sql.DB, notificationID, userID int) error {
	query := `
		DELETE FROM notifications
		WHERE id = ? AND user_id = ?
	`

	_, err := database.Exec(query, notificationID, userID)
	return err
}

// DeleteAllRead deletes all read notifications for a user
func DeleteAllRead(database *sql.DB, userID int) error {
	query := `
		DELETE FROM notifications
		WHERE user_id = ? AND is_read = 1
	`

	_, err := database.Exec(query, userID)
	return err
}
