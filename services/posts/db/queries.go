package db

import (
	"database/sql"
	"social-network/services/posts/models"
)

// CreatePost inserts a new post into the database
func CreatePost(db *sql.DB, post *models.Post) error {
	query := `
		INSERT INTO posts (user_id, title, content, image_path, privacy_level, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, post.UserID, post.Title, post.Content, post.ImagePath, post.PrivacyLevel, post.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	post.ID = int(id)
	return nil
}

// GetPostByID retrieves a post by ID
func GetPostByID(db *sql.DB, postID int) (*models.Post, error) {
	query := `
		SELECT id, user_id, title, content, image_path, privacy_level, created_at
		FROM posts
		WHERE id = ?
	`
	post := &models.Post{}
	err := db.QueryRow(query, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.ImagePath,
		&post.PrivacyLevel,
		&post.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// UpdatePost updates an existing post
func UpdatePost(db *sql.DB, post *models.Post) error {
	query := `
		UPDATE posts
		SET content = ?, image_path = ?, privacy_level = ?
		WHERE id = ?
	`
	_, err := db.Exec(query, post.Content, post.ImagePath, post.PrivacyLevel, post.ID)
	return err
}

// DeletePost deletes a post by ID
func DeletePost(db *sql.DB, postID int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := db.Exec(query, postID)
	return err
}

// GetPostsByUserID retrieves all posts by a specific user (for user's own profile)
func GetPostsByUserID(db *sql.DB, userID int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, title, content, image_path, privacy_level, created_at
		FROM posts
		WHERE user_id = ?
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
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// GetFeedPosts retrieves posts for a user's feed (public + following + own posts)
func GetFeedPosts(db *sql.DB, userID int) ([]*models.Post, error) {
	query := `
		SELECT DISTINCT 
			p.id, p.user_id, p.title, p.content, p.image_path, p.privacy_level, p.created_at,
			u.username, u.first_name, u.last_name, u.avatar_path
		FROM posts p
		INNER JOIN users u ON p.user_id = u.id
		LEFT JOIN follows f ON p.user_id = f.following_id AND f.follower_id = ? AND f.status = 'accepted'
		LEFT JOIN post_viewers pv ON p.id = pv.post_id AND pv.user_id = ?
		WHERE 
			p.privacy_level = 'public' OR
			p.user_id = ? OR
			(p.privacy_level = 'almost_private' AND pv.user_id IS NOT NULL) OR
			(p.privacy_level = 'private' AND f.follower_id IS NOT NULL)
		ORDER BY p.created_at DESC
	`
	rows, err := db.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var title, imagePath, username, firstName, lastName, avatar sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&title,
			&post.Content,
			&imagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
			&username,
			&firstName,
			&lastName,
			&avatar,
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

		// Add author information
		post.Author = &models.Author{
			ID:        post.UserID,
			Username:  username.String,
			FirstName: firstName.String,
			LastName:  lastName.String,
			Avatar:    avatar.String,
		}

		posts = append(posts, post)
	}
	return posts, nil
}

// CheckPostAccess checks if a user can view a specific post
func CheckPostAccess(db *sql.DB, postID, userID int) (bool, error) {
	query := `
		SELECT 
			CASE 
				WHEN p.user_id = ? THEN 1
				WHEN p.privacy_level = 'public' THEN 1
				WHEN p.privacy_level = 'private' AND EXISTS (
					SELECT 1 FROM follows WHERE follower_id = ? AND following_id = p.user_id AND status = 'accepted'
				) THEN 1
				WHEN p.privacy_level = 'almost_private' AND EXISTS (
					SELECT 1 FROM post_viewers WHERE post_id = ? AND user_id = ?
				) THEN 1
				ELSE 0
			END as has_access
		FROM posts p
		WHERE p.id = ?
	`
	var hasAccess int
	err := db.QueryRow(query, userID, userID, postID, userID, postID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}
	return hasAccess == 1, nil
}

// AddPostViewers adds users who can view an "almost_private" post
func AddPostViewers(db *sql.DB, postID int, userIDs []int) error {
	// First, clear existing viewers
	_, err := db.Exec(`DELETE FROM post_viewers WHERE post_id = ?`, postID)
	if err != nil {
		return err
	}

	// Insert new viewers
	if len(userIDs) == 0 {
		return nil
	}

	stmt, err := db.Prepare(`INSERT INTO post_viewers (post_id, user_id) VALUES (?, ?)`)
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
	return nil
}

// CreateComment inserts a new comment into the database
func CreateComment(db *sql.DB, comment *models.Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content, image_path, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query, comment.PostID, comment.UserID, comment.Content, comment.ImagePath, comment.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	comment.ID = int(id)
	return nil
}

// GetCommentsByPostID retrieves all comments for a specific post
func GetCommentsByPostID(db *sql.DB, postID int) ([]*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, image_path, created_at
		FROM comments
		WHERE post_id = ?
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
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.ImagePath,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
