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
		    nickname = COALESCE(?, nickname),
		    about_me = COALESCE(?, about_me),
		    is_public_profile = COALESCE(?, is_public_profile)
		WHERE id = ?
	`

	log.Printf("UpdateUserProfile: Executing query for user %d with values: firstName=%v, lastName=%v, dob=%v, nickname=%v, about=%v, isPublic=%v",
		userID, req.FirstName, req.LastName, req.DateOfBirth, req.Nickname, req.AboutMe, req.IsPublicProfile)

	result, err := db.Exec(query,
		req.FirstName,
		req.LastName,
		req.DateOfBirth,
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
func SearchUsers(db *sql.DB, searchTerm string) ([]*models.User, error) {
	query := `
		SELECT id, username, email, first_name, last_name, date_of_birth, avatar_path, 
		       nickname, about_me, is_public_profile, created_at
		FROM users 
		WHERE username LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR nickname LIKE ?
		LIMIT 50
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(query, searchPattern, searchPattern, searchPattern, searchPattern)
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
