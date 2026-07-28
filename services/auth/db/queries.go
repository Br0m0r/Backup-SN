package db

import (
	"database/sql"
	"errors"

	"social-network/services/auth/models"

	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUser(database *sql.DB, username, email, passwordHash, firstName, lastName, dateOfBirth string, nickname, aboutMe *string) (*models.User, error) {
	user := &models.User{
		Username: username, Email: email, FirstName: &firstName, LastName: &lastName,
		DateOfBirth: &dateOfBirth, Nickname: nickname, AboutMe: aboutMe,
		IsPublicProfile: true,
	}
	err := database.QueryRow(`
		INSERT INTO accounts (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, username, email, passwordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, errors.New("username or email already exists")
		}
		return nil, err
	}
	return user, nil
}

func scanUser(row *sql.Row) (*models.User, error) {
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return &user, err
}

func GetUserByEmail(database *sql.DB, email string) (*models.User, error) {
	return scanUser(database.QueryRow(`
		SELECT id, username, email, password_hash, created_at
		FROM accounts WHERE email = $1
	`, email))
}

func GetUserByID(database *sql.DB, userID int) (*models.User, error) {
	return scanUser(database.QueryRow(`
		SELECT id, username, email, password_hash, created_at
		FROM accounts WHERE id = $1
	`, userID))
}

func GetUserByUsername(database *sql.DB, username string) (*models.User, error) {
	return scanUser(database.QueryRow(`
		SELECT id, username, email, password_hash, created_at
		FROM accounts WHERE username = $1
	`, username))
}

func UserExistsByEmail(database *sql.DB, email string) (bool, error) {
	var exists bool
	err := database.QueryRow("SELECT EXISTS (SELECT 1 FROM accounts WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func UserExistsByUsername(database *sql.DB, username string) (bool, error) {
	var exists bool
	err := database.QueryRow("SELECT EXISTS (SELECT 1 FROM accounts WHERE username = $1)", username).Scan(&exists)
	return exists, err
}

func DeleteUser(database *sql.DB, userID int) error {
	_, err := database.Exec("DELETE FROM accounts WHERE id = $1", userID)
	return err
}
