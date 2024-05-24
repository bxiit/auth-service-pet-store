package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// App returns app by id.
func (s *Storage) App(ctx context.Context, id int) (App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = $1")
	if err != nil {
		return App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, id)

	var app App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return App{}, fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}

		return App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
