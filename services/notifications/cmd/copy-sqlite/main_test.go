package main

import (
	"database/sql"
	"testing"
	"time"
)

func TestChecksumIsStableAndSensitiveToNullableRelatedID(t *testing.T) {
	createdAt := time.Date(2026, time.July, 14, 12, 30, 0, 0, time.FixedZone("test", 2*60*60))
	records := []notificationRecord{{
		ID:        1,
		UserID:    2,
		Type:      "follow",
		RelatedID: sql.NullInt64{Int64: 3, Valid: true},
		Content:   "followed you",
		CreatedAt: createdAt,
	}}

	first, err := checksum(records)
	if err != nil {
		t.Fatalf("checksum returned an error: %v", err)
	}
	records[0].CreatedAt = createdAt.UTC()
	second, err := checksum(records)
	if err != nil {
		t.Fatalf("checksum returned an error: %v", err)
	}
	if first != second {
		t.Fatalf("equivalent instants produced different checksums: %s != %s", first, second)
	}

	records[0].RelatedID.Valid = false
	third, err := checksum(records)
	if err != nil {
		t.Fatalf("checksum returned an error: %v", err)
	}
	if first == third {
		t.Fatal("nullable related ID change did not affect checksum")
	}
}
