package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"sso/internal/data/models"
)

type UserInfoStorage struct {
	db *sql.DB
}

func (ui *UserInfoStorage) GetUser(ctx context.Context, id int) (user *models.User, err error) {
	const op = "data.storage.GetUser"
	fail := func(e error) error {
		return fmt.Errorf("%s: %w", op, e)
	}

	stmt, err := ui.db.Prepare(`
										SELECT * FROM sso.users
										WHERE id = $1
										`)
	if err != nil {
		return nil, fail(err)
	}

	var u models.User
	row := stmt.QueryRowContext(ctx, id)
	if errors.Is(row.Err(), sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}

	err = row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash.PlainText,
		&u.Role,
		&u.Activated,
	)
	if err != nil {
		return nil, fail(err)
	}

	return &u, nil
}
