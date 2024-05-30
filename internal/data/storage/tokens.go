package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/data/models"
	"time"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func (s *Storage) SaveToken(ctx context.Context, tokenPlainText string, userId int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"
	tokenHash := sha256.Sum256([]byte(tokenPlainText))
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	stmt, err := s.db.Prepare(`
								INSERT INTO tokens(hash, user_id, expiry) 
								VALUES ($1, $2, $3)
								`)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, fmt.Errorf("%s: %w", op, ErrTokenNotSaved)
		default:
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fail(err)
	}
	defer tx.Rollback()

	row := stmt.QueryRowContext(ctx, tokenHash[:], userId, time.Now().Add(time.Hour))

	if row.Err() != nil {
		return false, fail(err)
	}
	return true, nil
}

func (s *Storage) IsAuthenticated(ctx context.Context, tokenPlainText string) (bool, error) {
	const op = "storage.sqlite.IsAdmin"
	tokenHash := sha256.Sum256([]byte(tokenPlainText))
	stmt, err := s.db.Prepare(`
								SELECT users.id,
								       users.username,
								       users.email,
								       users.password_hash,
								       users.user_role,
								       users.activated
								           FROM users
								    INNER JOIN tokens t ON
								        users.id = t.user_id
								         WHERE t.hash = $1
								           AND t.expiry > $2
								`)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		default:
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	row := stmt.QueryRowContext(ctx, tokenHash[:], time.Now())

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return false, nil
	}

	var user models.User
	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.Hash,
		&user.Role,
		&user.Activated,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
