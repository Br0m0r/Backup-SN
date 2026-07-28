package db

import (
	"database/sql"
	"social-network/services/posts/models"
	"strconv"
	"strings"
)

// CreatePost inserts a new post into the database
func CreatePost(db *sql.DB, post *models.Post) error {
	query := `
		INSERT INTO posts (user_id, group_id, title, content, image_path, privacy_level, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return db.QueryRow(query, post.UserID, post.GroupID, post.Title, post.Content, post.ImagePath, post.PrivacyLevel, post.CreatedAt).Scan(&post.ID)
}

// GetPostByID retrieves a post by ID
func GetPostByID(db *sql.DB, postID int) (*models.Post, error) {
	query := `
		SELECT id, user_id, group_id, title, content, image_path, privacy_level, created_at
		FROM posts
		WHERE id = $1
	`
	post := &models.Post{}
	var groupID sql.NullInt64
	var title, imagePath sql.NullString

	err := db.QueryRow(query, postID).Scan(
		&post.ID,
		&post.UserID,
		&groupID,
		&title,
		&post.Content,
		&imagePath,
		&post.PrivacyLevel,
		&post.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if groupID.Valid {
		gid := int(groupID.Int64)
		post.GroupID = &gid
	}
	if title.Valid {
		post.Title = &title.String
	}
	if imagePath.Valid {
		post.ImagePath = &imagePath.String
	}

	return post, nil
}

// UpdatePost updates an existing post
func UpdatePost(db *sql.DB, post *models.Post) error {
	query := `
		UPDATE posts
		SET content = $1, image_path = $2, privacy_level = $3
		WHERE id = $4
	`
	_, err := db.Exec(query, post.Content, post.ImagePath, post.PrivacyLevel, post.ID)
	return err
}

// DeletePost deletes a post by ID
func DeletePost(db *sql.DB, postID int) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := db.Exec(query, postID)
	return err
}

// GetPostsByUserID retrieves all posts by a specific user (for user's own profile)
func GetPostsByUserID(db *sql.DB, userID int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, group_id, title, content, image_path, privacy_level, created_at
		FROM posts
		WHERE user_id = $1 AND group_id IS NULL
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var groupID sql.NullInt64
		var title, content, imagePath sql.NullString
		var privacyLevel string
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&groupID,
			&title,
			&content,
			&imagePath,
			&privacyLevel,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if groupID.Valid {
			gid := int(groupID.Int64)
			post.GroupID = &gid
		}
		if title.Valid {
			post.Title = &title.String
		}
		if content.Valid {
			post.Content = content.String
		}
		if imagePath.Valid {
			post.ImagePath = &imagePath.String
		}
		post.PrivacyLevel = privacyLevel
		posts = append(posts, post)
	}
	return posts, nil
}

// GetFeedPosts retrieves posts for a user's feed (public + following + own posts)
func GetFeedPosts(db *sql.DB, userID int, followingIDs []int) ([]*models.Post, error) {
	followingClause, followingArguments := idClause("p.user_id", followingIDs, 2)
	viewerPlaceholder := "$" + strconv.Itoa(2+len(followingArguments))
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.image_path, p.privacy_level, p.created_at
		FROM posts p
		WHERE p.group_id IS NULL AND (
			p.privacy_level = 'public'
			OR p.user_id = $1
			OR (p.privacy_level = 'almost_private' AND ` + followingClause + `)
			OR (p.privacy_level = 'private' AND EXISTS (
				SELECT 1 FROM post_viewers WHERE post_id = p.id AND user_id = ` + viewerPlaceholder + `
			))
		)
		ORDER BY p.created_at DESC
	`
	arguments := []any{userID}
	arguments = append(arguments, followingArguments...)
	arguments = append(arguments, userID)
	rows, err := db.Query(query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var title, imagePath sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&title,
			&post.Content,
			&imagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if title.Valid {
			post.Title = &title.String
		}
		if imagePath.Valid {
			post.ImagePath = &imagePath.String
		}

		posts = append(posts, post)
	}
	return posts, nil
}

// GetPostsByGroupID retrieves all posts for a specific group
func GetPostsByGroupID(db *sql.DB, groupID int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, group_id, title, content, image_path, privacy_level, created_at
		FROM posts
		WHERE group_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var postGroupID sql.NullInt64
		var title, imagePath sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&postGroupID,
			&title,
			&post.Content,
			&imagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if postGroupID.Valid {
			gid := int(postGroupID.Int64)
			post.GroupID = &gid
		}
		if title.Valid {
			post.Title = &title.String
		}
		if imagePath.Valid {
			post.ImagePath = &imagePath.String
		}

		posts = append(posts, post)
	}
	return posts, nil
}

// SearchPosts searches for posts based on query string (searches content and title)
func SearchPosts(db *sql.DB, userID int, searchQuery string, followingIDs, authorIDs []int) ([]*models.Post, error) {
	followingClause, followingArguments := idClause("p.user_id", followingIDs, 2)
	viewerIndex := 2 + len(followingArguments)
	searchIndex := viewerIndex + 1
	authorClause, authorArguments := idClause("p.user_id", authorIDs, searchIndex+2)
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.image_path, p.privacy_level, p.created_at
		FROM posts p
		WHERE p.group_id IS NULL AND (
			p.privacy_level = 'public'
			OR p.user_id = $1
			OR (p.privacy_level = 'almost_private' AND ` + followingClause + `)
			OR (p.privacy_level = 'private' AND EXISTS (
				SELECT 1 FROM post_viewers WHERE post_id = p.id AND user_id = $` + strconv.Itoa(viewerIndex) + `
			))
		) AND (
			p.content ILIKE $` + strconv.Itoa(searchIndex) + `
			OR p.title ILIKE $` + strconv.Itoa(searchIndex+1) + `
			OR ` + authorClause + `
		)
		ORDER BY p.created_at DESC
	`
	searchPattern := "%" + searchQuery + "%"
	arguments := []any{userID}
	arguments = append(arguments, followingArguments...)
	arguments = append(arguments, userID, searchPattern, searchPattern)
	arguments = append(arguments, authorArguments...)
	rows, err := db.Query(query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var title, imagePath sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&title,
			&post.Content,
			&imagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if title.Valid {
			post.Title = &title.String
		}
		if imagePath.Valid {
			post.ImagePath = &imagePath.String
		}

		posts = append(posts, post)
	}
	return posts, nil
}

// CheckPostAccess checks if a user can view a specific post
func CheckPostAccess(db *sql.DB, postID, userID int, followsAuthor bool) (bool, error) {
	var authorID int
	var privacyLevel string
	err := db.QueryRow(`SELECT user_id, privacy_level FROM posts WHERE id = $1`, postID).Scan(&authorID, &privacyLevel)
	if err != nil {
		return false, err
	}
	if authorID == userID || privacyLevel == "public" || (privacyLevel == "almost_private" && followsAuthor) {
		return true, nil
	}
	if privacyLevel != "private" {
		return false, nil
	}
	var viewerCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM post_viewers WHERE post_id = $1 AND user_id = $2`, postID, userID).Scan(&viewerCount)
	return viewerCount > 0, err
}

// AddPostViewers adds users who can view a "private" post
func AddPostViewers(db *sql.DB, postID int, userIDs []int) error {
	// First, clear existing viewers
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err := transaction.Exec(`DELETE FROM post_viewers WHERE post_id = $1`, postID); err != nil {
		return err
	}

	// Insert new viewers
	if len(userIDs) == 0 {
		return transaction.Commit()
	}

	stmt, err := transaction.Prepare(`
		INSERT INTO post_viewers (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.Exec(postID, userID)
		if err != nil {
			return err
		}
	}
	return transaction.Commit()
}

// CreateComment inserts a new comment into the database
func CreateComment(db *sql.DB, comment *models.Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content, image_path, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return db.QueryRow(query, comment.PostID, comment.UserID, comment.Content, comment.ImagePath, comment.CreatedAt).Scan(&comment.ID)
}

// GetCommentsByPostID retrieves all comments for a specific post
func GetCommentsByPostID(db *sql.DB, postID int) ([]*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, image_path, created_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC
	`
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		var imagePath sql.NullString

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&imagePath,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if imagePath.Valid {
			comment.ImagePath = &imagePath.String
		}

		comments = append(comments, comment)
	}
	return comments, nil
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(db *sql.DB, commentID int) (*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, image_path, created_at
		FROM comments
		WHERE id = $1
	`
	comment := &models.Comment{}
	var imagePath sql.NullString

	err := db.QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&imagePath,
		&comment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if imagePath.Valid {
		comment.ImagePath = &imagePath.String
	}

	return comment, nil
}

// UpdateComment updates an existing comment
func UpdateComment(db *sql.DB, comment *models.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, image_path = $2
		WHERE id = $3
	`
	_, err := db.Exec(query, comment.Content, comment.ImagePath, comment.ID)
	return err
}

// DeleteComment deletes a comment by ID
func DeleteComment(db *sql.DB, commentID int) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := db.Exec(query, commentID)
	return err
}

func idClause(column string, userIDs []int, start int) (string, []any) {
	if len(userIDs) == 0 {
		return "1 = 0", nil
	}
	placeholders := make([]string, len(userIDs))
	arguments := make([]any, len(userIDs))
	for index, userID := range userIDs {
		placeholders[index] = "$" + strconv.Itoa(start+index)
		arguments[index] = userID
	}
	return column + " IN (" + strings.Join(placeholders, ",") + ")", arguments
}
