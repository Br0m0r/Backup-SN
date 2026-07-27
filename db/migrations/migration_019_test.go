package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestAddGroupActivityNotificationMigration(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })

	usersMigration := readMigration(t, "000001_Users.up.sql")
	notificationsMigration := readMigration(t, "000007_Notifications.up.sql")
	if _, err := database.Exec(usersMigration + "\n" + notificationsMigration); err != nil {
		t.Fatalf("create original schema: %v", err)
	}
	if _, err := database.Exec(`
		INSERT INTO users (username, email, password_hash) VALUES ('alice', 'alice@example.com', 'hash');
		INSERT INTO notifications (user_id, type, content) VALUES (1, 'group_request', 'existing');
	`); err != nil {
		t.Fatalf("create original data: %v", err)
	}

	if _, err := database.Exec(readMigration(t, "000019_AddGroupActivityNotification.up.sql")); err != nil {
		t.Fatalf("apply migration up: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO notifications (user_id, type, content) VALUES (1, 'group_activity', 'joined')`); err != nil {
		t.Fatalf("new notification type was rejected: %v", err)
	}

	if _, err := database.Exec(readMigration(t, "000019_AddGroupActivityNotification.down.sql")); err != nil {
		t.Fatalf("apply migration down: %v", err)
	}

	var mappedCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications WHERE type = 'group_request'`).Scan(&mappedCount); err != nil {
		t.Fatalf("count mapped notifications: %v", err)
	}
	if mappedCount != 2 {
		t.Fatalf("group_request count after rollback = %d, want 2", mappedCount)
	}
	if _, err := database.Exec(`INSERT INTO notifications (user_id, type, content) VALUES (1, 'group_activity', 'invalid')`); err == nil {
		t.Fatal("rollback did not restore the original notification type constraint")
	}
}

func readMigration(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(contents)
}
