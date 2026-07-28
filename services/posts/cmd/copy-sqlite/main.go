package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"social-network/services/common/postgres"
	"social-network/services/posts/migrations"

	_ "github.com/mattn/go-sqlite3"
)

type postRecord struct {
	ID, UserID            int64
	GroupID               sql.NullInt64
	Title, ImagePath      sql.NullString
	Content, PrivacyLevel string
	CreatedAt             time.Time
}
type viewerRecord struct {
	ID, PostID, UserID int64
	CreatedAt          time.Time
}
type commentRecord struct {
	ID, PostID, UserID int64
	Content            string
	ImagePath          sql.NullString
	CreatedAt          time.Time
}
type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func main() {
	sqlitePath := flag.String("sqlite-path", "./social_network.db", "legacy SQLite path")
	flag.Parse()
	if err := run(*sqlitePath); err != nil {
		log.Fatal(err)
	}
}

func run(sqlitePath string) error {
	if _, err := os.Stat(sqlitePath); err != nil {
		return err
	}
	config, err := postgres.FromEnvironment()
	if err != nil {
		return err
	}
	target, err := postgres.Open(context.Background(), config)
	if err != nil {
		return err
	}
	defer target.Close()
	if err := migrations.Apply(context.Background(), target); err != nil {
		return err
	}
	source, err := sql.Open("sqlite3", "file:"+sqlitePath+"?mode=ro")
	if err != nil {
		return err
	}
	defer source.Close()
	ctx := context.Background()
	posts, viewers, comments, err := readAll(ctx, source)
	if err != nil {
		return err
	}
	transaction, err := target.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	for _, row := range posts {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO posts (id,user_id,group_id,title,content,image_path,privacy_level,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (id) DO UPDATE SET user_id=EXCLUDED.user_id,
				group_id=EXCLUDED.group_id,title=EXCLUDED.title,content=EXCLUDED.content,
				image_path=EXCLUDED.image_path,privacy_level=EXCLUDED.privacy_level,
				created_at=EXCLUDED.created_at`,
			row.ID, row.UserID, row.GroupID, row.Title, row.Content, row.ImagePath, row.PrivacyLevel, row.CreatedAt); err != nil {
			return fmt.Errorf("copy post %d: %w", row.ID, err)
		}
	}
	for _, row := range viewers {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO post_viewers (id,post_id,user_id,created_at) VALUES ($1,$2,$3,$4)
			ON CONFLICT (id) DO UPDATE SET post_id=EXCLUDED.post_id,
				user_id=EXCLUDED.user_id,created_at=EXCLUDED.created_at`,
			row.ID, row.PostID, row.UserID, row.CreatedAt); err != nil {
			return fmt.Errorf("copy post viewer %d: %w", row.ID, err)
		}
	}
	for _, row := range comments {
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO comments (id,post_id,user_id,content,image_path,created_at)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id) DO UPDATE SET post_id=EXCLUDED.post_id,
				user_id=EXCLUDED.user_id,content=EXCLUDED.content,
				image_path=EXCLUDED.image_path,created_at=EXCLUDED.created_at`,
			row.ID, row.PostID, row.UserID, row.Content, row.ImagePath, row.CreatedAt); err != nil {
			return fmt.Errorf("copy comment %d: %w", row.ID, err)
		}
	}
	targetPosts, targetViewers, targetComments, err := readAll(ctx, transaction)
	if err != nil {
		return err
	}
	sourceChecksum, err := checksum(posts, viewers, comments)
	if err != nil {
		return err
	}
	targetChecksum, err := checksum(targetPosts, targetViewers, targetComments)
	if err != nil {
		return err
	}
	if len(posts) != len(targetPosts) || len(viewers) != len(targetViewers) ||
		len(comments) != len(targetComments) || sourceChecksum != targetChecksum {
		return fmt.Errorf("Posts verification failed; target transaction rolled back")
	}
	for _, table := range []string{"posts", "post_viewers", "comments"} {
		if _, err := transaction.ExecContext(ctx, fmt.Sprintf(`
			SELECT setval(pg_get_serial_sequence('%s','id')::regclass,
				COALESCE(MAX(id),1),COUNT(*) > 0) FROM %s`, table, table)); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	log.Printf("Copied %d posts, %d viewers, and %d comments (checksum %s)", len(posts), len(viewers), len(comments), sourceChecksum)
	return nil
}

func readAll(ctx context.Context, database queryer) ([]postRecord, []viewerRecord, []commentRecord, error) {
	posts := make([]postRecord, 0)
	rows, err := database.QueryContext(ctx, `SELECT id,user_id,group_id,title,content,image_path,privacy_level,created_at FROM posts ORDER BY id`)
	if err != nil {
		return nil, nil, nil, err
	}
	for rows.Next() {
		var row postRecord
		if err := rows.Scan(&row.ID, &row.UserID, &row.GroupID, &row.Title, &row.Content, &row.ImagePath, &row.PrivacyLevel, &row.CreatedAt); err != nil {
			rows.Close()
			return nil, nil, nil, err
		}
		posts = append(posts, row)
	}
	if err := rows.Close(); err != nil {
		return nil, nil, nil, err
	}
	viewers := make([]viewerRecord, 0)
	rows, err = database.QueryContext(ctx, `SELECT id,post_id,user_id,created_at FROM post_viewers ORDER BY id`)
	if err != nil {
		return nil, nil, nil, err
	}
	for rows.Next() {
		var row viewerRecord
		if err := rows.Scan(&row.ID, &row.PostID, &row.UserID, &row.CreatedAt); err != nil {
			rows.Close()
			return nil, nil, nil, err
		}
		viewers = append(viewers, row)
	}
	if err := rows.Close(); err != nil {
		return nil, nil, nil, err
	}
	comments := make([]commentRecord, 0)
	rows, err = database.QueryContext(ctx, `SELECT id,post_id,user_id,content,image_path,created_at FROM comments ORDER BY id`)
	if err != nil {
		return nil, nil, nil, err
	}
	for rows.Next() {
		var row commentRecord
		if err := rows.Scan(&row.ID, &row.PostID, &row.UserID, &row.Content, &row.ImagePath, &row.CreatedAt); err != nil {
			rows.Close()
			return nil, nil, nil, err
		}
		comments = append(comments, row)
	}
	return posts, viewers, comments, rows.Close()
}

func checksum(posts []postRecord, viewers []viewerRecord, comments []commentRecord) (string, error) {
	hash := sha256.New()
	encoder := json.NewEncoder(hash)
	posts = append([]postRecord(nil), posts...)
	viewers = append([]viewerRecord(nil), viewers...)
	comments = append([]commentRecord(nil), comments...)
	for index := range posts {
		posts[index].CreatedAt = posts[index].CreatedAt.UTC()
	}
	for index := range viewers {
		viewers[index].CreatedAt = viewers[index].CreatedAt.UTC()
	}
	for index := range comments {
		comments[index].CreatedAt = comments[index].CreatedAt.UTC()
	}
	for _, records := range []any{posts, viewers, comments} {
		if err := encoder.Encode(records); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
