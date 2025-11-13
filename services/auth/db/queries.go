package db

import (
	"database/sql"
	"errors"
	"time"

	"social-network/services/auth/models"

	"github.com/mattn/go-sqlite3"
)

// CreateUser inserts a new user into the database
func CreateUser(db *sql.DB, username, email, passwordHash, firstName, lastName string) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, first_name, last_name, is_public_profile, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := db.Exec(query, username, email, passwordHash, firstName, lastName, true, now)
	if err != nil {
		// Check if it's a unique constraint violation (email or username already exists)
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return nil, errors.New("username or email already exists")
		}
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:              int(userID),
		Username:        username,
		Email:           email,
		FirstName:       &firstName,
		LastName:        &lastName,
		IsPublicProfile: true,
		CreatedAt:       now,
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email address
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, date_of_birth, avatar_path, 
		       nickname, about_me, is_public_profile, created_at
		FROM users 
		WHERE email = ?
	`

	var user models.User
	var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
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

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, userID int) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, date_of_birth, avatar_path, 
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
		&user.PasswordHash,
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

// UserExistsByEmail checks if a user exists with the given email
func UserExistsByEmail(db *sql.DB, email string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserExistsByUsername checks if a user exists with the given username
func UserExistsByUsername(db *sql.DB, username string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(db *sql.DB, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, date_of_birth, avatar_path, 
		       nickname, about_me, is_public_profile, created_at
		FROM users 
		WHERE username = ?
	`

	var user models.User
	var firstName, lastName, dateOfBirth, avatarPath, nickname, aboutMe sql.NullString

	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
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
