package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"social-network/services/chat/migrations"
	"social-network/services/chat/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestMessageRepositoryAgainstPostgres(t *testing.T) {
	databaseURL := os.Getenv("CHAT_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CHAT_TEST_DATABASE_URL is not configured")
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
	if _, err := database.Exec("TRUNCATE TABLE messages, group_messages RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate Chat tables: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	message := &models.Message{SenderID: 1, ReceiverID: 2, Content: "hello", CreatedAt: now}
	if err := SaveMessage(database, message); err != nil || message.ID != 1 {
		t.Fatalf("SaveMessage() id=%d error=%v", message.ID, err)
	}
	history, err := GetChatHistory(database, 1, 2, 20)
	if err != nil || len(history) != 1 || history[0].Content != "hello" {
		t.Fatalf("GetChatHistory()=%+v error=%v", history, err)
	}
	if count, err := GetUnreadCount(database, 2); err != nil || count != 1 {
		t.Fatalf("GetUnreadCount()=%d error=%v", count, err)
	}
	if err := MarkAsRead(database, 1, 2); err != nil {
		t.Fatalf("MarkAsRead(): %v", err)
	}

	groupMessage := &models.GroupMessage{GroupID: 7, SenderID: 1, Content: "group", CreatedAt: now}
	if err := SaveGroupMessage(database, groupMessage); err != nil || groupMessage.ID != 1 {
		t.Fatalf("SaveGroupMessage() id=%d error=%v", groupMessage.ID, err)
	}
	groupHistory, err := GetGroupChatHistory(database, 7, 20)
	if err != nil || len(groupHistory) != 1 || groupHistory[0].Content != "group" {
		t.Fatalf("GetGroupChatHistory()=%+v error=%v", groupHistory, err)
	}
}
