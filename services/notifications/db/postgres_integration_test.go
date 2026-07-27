package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"social-network/services/notifications/migrations"
	"social-network/services/notifications/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNotificationRepositoryAgainstPostgres(t *testing.T) {
	databaseURL := os.Getenv("NOTIFICATIONS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("NOTIFICATIONS_TEST_DATABASE_URL is not configured")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	defer database.Close()
	if err := database.PingContext(context.Background()); err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}
	if err := migrations.Apply(context.Background(), database); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if _, err := database.Exec("TRUNCATE TABLE notifications RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate notifications: %v", err)
	}
	t.Cleanup(func() {
		_, _ = database.Exec("TRUNCATE TABLE notifications RESTART IDENTITY")
	})

	created, err := CreateNotification(database, &models.CreateNotificationRequest{
		UserID:    42,
		Type:      models.TypeGroupActivity,
		RelatedID: 7,
		Content:   "joined your group",
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}
	if created.ID != 1 || created.IsRead || created.Type != models.TypeGroupActivity {
		t.Fatalf("unexpected created notification: %+v", created)
	}

	fetched, err := GetNotificationByID(database, created.ID)
	if err != nil {
		t.Fatalf("GetNotificationByID: %v", err)
	}
	if fetched.Content != created.Content || fetched.UserID != 42 {
		t.Fatalf("unexpected fetched notification: %+v", fetched)
	}

	count, err := GetUnreadCount(database, 42)
	if err != nil || count != 1 {
		t.Fatalf("GetUnreadCount = %d, %v; want 1, nil", count, err)
	}
	if err := MarkAsRead(database, created.ID, 42); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}
	count, err = GetUnreadCount(database, 42)
	if err != nil || count != 0 {
		t.Fatalf("GetUnreadCount after read = %d, %v; want 0, nil", count, err)
	}

	notifications, err := GetUserNotifications(database, 42, 20, 0)
	if err != nil || len(notifications) != 1 {
		t.Fatalf("GetUserNotifications length = %d, error = %v", len(notifications), err)
	}
	if err := DeleteNotification(database, created.ID, 42); err != nil {
		t.Fatalf("DeleteNotification: %v", err)
	}
	if _, err := GetNotificationByID(database, created.ID); err != sql.ErrNoRows {
		t.Fatalf("GetNotificationByID after delete = %v; want sql.ErrNoRows", err)
	}
}
