package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"sso/internal/data/models"
)

type AuthStorage struct {
	db *sql.DB
}

func (s *AuthStorage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "data.storage.SaveUser"
	fail := func(e error) error {
		return fmt.Errorf("%s: %w", op, e)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return 0, fail(err)
	}

	stmt, err := s.db.Prepare("INSERT INTO sso.users(username, email, password_hash, user_role, activated) VALUES($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		return 0, fail(err)
	}

	//res, err := stmt.ExecContext(ctx, "default username", email, passHash, "user", false)
	var id int64
	err = stmt.QueryRowContext(ctx, "default username", email, passHash, "user", false).Scan(&id)
	if err != nil {
		return 0, fail(err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fail(err)
	}

	return id, nil
}

// User returns user by email.
func (s *AuthStorage) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, username, email, password_hash, user_role, activated FROM sso.users WHERE email = $1")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash.Hash, &user.Role, &user.Activated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *AuthStorage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"
	stmt, err := s.db.Prepare("SELECT user_role FROM sso.users WHERE id = $1")
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		default:
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	row := stmt.QueryRowContext(ctx, userID)

	var role string
	err = row.Scan(&role)
	if err != nil {
		println(fmt.Errorf("%s: %w", op, ErrUserNotFound))
		return false, err
	}

	isRoleAdmin := role == "admin"
	return isRoleAdmin, nil
}
