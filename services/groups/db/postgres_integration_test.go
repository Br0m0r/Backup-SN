package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"social-network/services/groups/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestGroupRepositoryAgainstPostgres(t *testing.T) {
	databaseURL := os.Getenv("GROUPS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("GROUPS_TEST_DATABASE_URL is not configured")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.PingContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Apply(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("TRUNCATE TABLE event_responses, events, group_members, groups RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}

	group, err := CreateGroup(database, "PostgreSQL Group", nil, nil, 7)
	if err != nil || group.ID != 1 {
		t.Fatalf("CreateGroup()=%v, %v", group, err)
	}
	if err := RequestToJoinGroup(database, group.ID, 8); err != nil {
		t.Fatal(err)
	}
	participants, err := GetGroupParticipantIDs(database, group.ID)
	if err != nil || len(participants) != 2 {
		t.Fatalf("GetGroupParticipantIDs()=%v, %v", participants, err)
	}
	event, err := CreateEvent(database, group.ID, 7, "Launch", nil, time.Now().UTC().Add(time.Hour))
	if err != nil || event.ID != 1 {
		t.Fatalf("CreateEvent()=%v, %v", event, err)
	}
	if err := RespondToEvent(database, event.ID, 7, "going"); err != nil {
		t.Fatal(err)
	}
	withResponses, err := GetEventWithResponses(database, event.ID, 7)
	if err != nil || withResponses.GoingCount != 1 || withResponses.UserResponse != "going" {
		t.Fatalf("GetEventWithResponses()=%v, %v", withResponses, err)
	}
}
