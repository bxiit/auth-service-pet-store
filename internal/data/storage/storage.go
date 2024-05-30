package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrAppNotFound   = errors.New("app not found")
	ErrTokenNotSaved = errors.New("token not saved")
)

func New(dsn string) (*Storage, error) {
	const op = "data.sqlite.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop(db *sql.DB) error {
	return s.db.Close()
}
