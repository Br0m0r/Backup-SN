package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"social-network/services/posts/migrations"
	"social-network/services/posts/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostRepositoryAgainstPostgres(t *testing.T) {
	databaseURL := os.Getenv("POSTS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("POSTS_TEST_DATABASE_URL is not configured")
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
	if _, err := database.Exec("TRUNCATE TABLE comments, post_viewers, posts RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	post := &models.Post{UserID: 7, Content: "hello", PrivacyLevel: "almost_private", CreatedAt: now}
	if err := CreatePost(database, post); err != nil || post.ID != 1 {
		t.Fatalf("CreatePost() id=%d error=%v", post.ID, err)
	}
	feed, err := GetFeedPosts(database, 42, []int{7})
	if err != nil || len(feed) != 1 {
		t.Fatalf("GetFeedPosts()=%v, %v", feed, err)
	}
	search, err := SearchPosts(database, 42, "hell", []int{7}, nil)
	if err != nil || len(search) != 1 {
		t.Fatalf("SearchPosts()=%v, %v", search, err)
	}
	comment := &models.Comment{PostID: post.ID, UserID: 42, Content: "reply", CreatedAt: now}
	if err := CreateComment(database, comment); err != nil || comment.ID != 1 {
		t.Fatalf("CreateComment() id=%d error=%v", comment.ID, err)
	}
	comments, err := GetCommentsByPostID(database, post.ID)
	if err != nil || len(comments) != 1 {
		t.Fatalf("GetCommentsByPostID()=%v, %v", comments, err)
	}
}
