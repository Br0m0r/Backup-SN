package db

import (
	"database/sql"
	"errors"
	"log"

	"social-network/services/users/models"
)

// GetUserByID retrieves a user profile by ID
func GetUserByID(db *sql.DB, userID int) (*models.User, error) {
	query := `
		SELECT id, username, email, first_name, last_name, date_of_birth, avatar_path, 
		       nickname, about_me, is_public_profile, created_at
		FROM users 
		WHERE id = ?
	`

	var user models.User
	var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

	err := db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&firstName,
		&lastName,
		&dateOfBirth,
		&avatarPath,
		&nickname,
		&aboutMe,
		&user.IsPublicProfile,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Handle nullable fields
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if dateOfBirth.Valid {
		user.DateOfBirth = &dateOfBirth.String
	}
	if avatarPath.Valid {
		user.AvatarPath = &avatarPath.String
	}
	if nickname.Valid {
		user.Nickname = &nickname.String
	}
	if aboutMe.Valid {
		user.AboutMe = &aboutMe.String
	}

	return &user, nil
}

// UpdateUserProfile updates user profile information
func UpdateUserProfile(db *sql.DB, userID int, req *models.UpdateProfileRequest) error {
	query := `
		UPDATE users 
		SET first_name = COALESCE(?, first_name),
		    last_name = COALESCE(?, last_name),
		    date_of_birth = COALESCE(?, date_of_birth),
		    avatar_path = COALESCE(?, avatar_path),
		    nickname = COALESCE(?, nickname),
		    about_me = COALESCE(?, about_me),
		    is_public_profile = COALESCE(?, is_public_profile)
		WHERE id = ?
	`

	log.Printf("UpdateUserProfile: Executing query for user %d with values: firstName=%v, lastName=%v, dob=%v, avatar=%v, nickname=%v, about=%v, isPublic=%v",
		userID, req.FirstName, req.LastName, req.DateOfBirth, req.AvatarPath, req.Nickname, req.AboutMe, req.IsPublicProfile)

	result, err := db.Exec(query,
		req.FirstName,
		req.LastName,
		req.DateOfBirth,
		req.AvatarPath,
		req.Nickname,
		req.AboutMe,
		req.IsPublicProfile,
		userID,
	)

	if err != nil {
		log.Printf("UpdateUserProfile: Database error: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("UpdateUserProfile: Updated %d rows for user %d", rowsAffected, userID)

	return nil
}

// CreateFollow creates a follow relationship
func CreateFollow(db *sql.DB, followerID, followingID int, status string) error {
	query := `
		INSERT INTO follows (follower_id, following_id, status, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`

	_, err := db.Exec(query, followerID, followingID, status)
	if err != nil {
		return errors.New("failed to create follow relationship")
	}

	return nil
}

// DeleteFollow removes a follow relationship
func DeleteFollow(db *sql.DB, followerID, followingID int) error {
	query := `DELETE FROM follows WHERE follower_id = ? AND following_id = ?`

	result, err := db.Exec(query, followerID, followingID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("follow relationship not found")
	}

	return nil
}

// GetFollowers retrieves all followers of a user
func GetFollowers(db *sql.DB, userID int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.date_of_birth, 
		       u.avatar_path, u.nickname, u.about_me, u.is_public_profile, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, &user)
	}

	return users, nil
}

// GetFollowing retrieves all users that a user is following
func GetFollowing(db *sql.DB, userID int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.date_of_birth, 
		       u.avatar_path, u.nickname, u.about_me, u.is_public_profile, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.following_id
		WHERE f.follower_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, &user)
	}

	return users, nil
}

// SearchUsers searches for users by username or name
func SearchUsers(db *sql.DB, searchTerm string, currentUserID int) ([]*models.User, error) {
	query := `
		SELECT DISTINCT u.id, u.username, u.email, u.first_name, u.last_name, u.date_of_birth, u.avatar_path, 
		       u.nickname, u.about_me, u.is_public_profile, u.created_at
		FROM users u
		WHERE (u.username LIKE ? OR u.first_name LIKE ? OR u.last_name LIKE ? OR u.nickname LIKE ?)
		  AND u.id != ?
		  AND u.id NOT IN (
		    SELECT following_id FROM follows 
		    WHERE follower_id = ? AND status IN ('accepted', 'pending')
		  )
		LIMIT 50
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(query, searchPattern, searchPattern, searchPattern, searchPattern, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, &user)
	}

	return users, nil
}

// SearchUsersForGroup searches for users to invite to a group (excludes only current group members)
func SearchUsersForGroup(db *sql.DB, searchTerm string, currentUserID int, groupID int) ([]*models.User, error) {
	query := `
		SELECT DISTINCT u.id, u.username, u.email, u.first_name, u.last_name, u.date_of_birth, u.avatar_path, 
		       u.nickname, u.about_me, u.is_public_profile, u.created_at
		FROM users u
		WHERE (u.username LIKE ? OR u.first_name LIKE ? OR u.last_name LIKE ? OR u.nickname LIKE ?)
		  AND u.id != ?
		  AND u.id NOT IN (
		    SELECT user_id FROM group_members 
		    WHERE group_id = ?
		  )
		LIMIT 50
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(query, searchPattern, searchPattern, searchPattern, searchPattern, currentUserID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, &user)
	}

	return users, nil
}

// CheckFollowStatus checks if a follow relationship exists and its status
func CheckFollowStatus(db *sql.DB, followerID, followingID int) (string, error) {
	query := `SELECT status FROM follows WHERE follower_id = ? AND following_id = ?`

	var status string
	err := db.QueryRow(query, followerID, followingID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "none", nil
		}
		return "", err
	}

	return status, nil
}

// GetPendingFollowRequests retrieves all pending follow requests for a user
func GetPendingFollowRequests(db *sql.DB, userID int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.date_of_birth, 
		       u.avatar_path, u.nickname, u.about_me, u.is_public_profile, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = ? AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, &user)
	}

	return users, rows.Err()
}

// RespondToFollowRequest accepts or rejects a follow request
func RespondToFollowRequest(db *sql.DB, followerID, followingID int, accept bool) error {
	if accept {
		// Update status to accepted
		query := `UPDATE follows SET status = 'accepted' WHERE follower_id = ? AND following_id = ? AND status = 'pending'`
		result, err := db.Exec(query, followerID, followingID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return errors.New("follow request not found or already processed")
		}

		return nil
	} else {
		// Delete the pending request
		query := `DELETE FROM follows WHERE follower_id = ? AND following_id = ? AND status = 'pending'`
		result, err := db.Exec(query, followerID, followingID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return errors.New("follow request not found")
		}

		return nil
	}
}

// GetUserPosts retrieves all posts by a user from the posts database
func GetUserPosts(db *sql.DB, userID int) ([]models.UserPost, error) {
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

	var posts []models.UserPost
	for rows.Next() {
		var post models.UserPost
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
			log.Printf("Error scanning post: %v", err)
			continue
		}

		if title.Valid {
			post.Title = &title.String
		}
		if imagePath.Valid {
			post.ImagePath = &imagePath.String
		}

		posts = append(posts, post)
	}

	if posts == nil {
		posts = []models.UserPost{}
	}

	return posts, nil
}

// GetUserFollowersList retrieves followers with accepted status
func GetUserFollowersList(db *sql.DB, userID int) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.first_name, u.last_name, 
		       u.date_of_birth, u.avatar_path, u.nickname, u.about_me, 
		       u.is_public_profile, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning follower: %v", err)
			continue
		}

		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	return users, nil
}

// GetUserFollowingList retrieves users that the given user is following (accepted status)
func GetUserFollowingList(db *sql.DB, userID int) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.first_name, u.last_name, 
		       u.date_of_birth, u.avatar_path, u.nickname, u.about_me, 
		       u.is_public_profile, u.created_at
		FROM users u
		INNER JOIN follows f ON u.id = f.following_id
		WHERE f.follower_id = ? AND f.status = 'accepted'
		ORDER BY f.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&firstName,
			&lastName,
			&dateOfBirth,
			&avatarPath,
			&nickname,
			&aboutMe,
			&user.IsPublicProfile,
			&user.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning following: %v", err)
			continue
		}

		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if dateOfBirth.Valid {
			user.DateOfBirth = &dateOfBirth.String
		}
		if avatarPath.Valid {
			user.AvatarPath = &avatarPath.String
		}
		if nickname.Valid {
			user.Nickname = &nickname.String
		}
		if aboutMe.Valid {
			user.AboutMe = &aboutMe.String
		}

		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	return users, nil
}

// CheckProfileAccess checks if viewerID can access userID's profile
// Returns true if: viewer is owner, profile is public, or viewer follows user (for private profiles)
func CheckProfileAccess(db *sql.DB, userID, viewerID int) (bool, error) {
	// Owner can always see their own profile
	if userID == viewerID {
		return true, nil
	}

	// Check if profile is public
	var isPublic bool
	err := db.QueryRow(`SELECT is_public_profile FROM users WHERE id = ?`, userID).Scan(&isPublic)
	if err != nil {
		return false, err
	}

	// If public, everyone can see
	if isPublic {
		return true, nil
	}

	// If private, only followers (accepted status) can see
	var count int
	query := `SELECT COUNT(*) FROM follows WHERE follower_id = ? AND following_id = ? AND status = 'accepted'`
	err = db.QueryRow(query, viewerID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
