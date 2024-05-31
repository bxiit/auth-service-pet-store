package user_info

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/data/models"
	"sso/internal/services/auth"
	"time"
)

type UserInfoProvider interface {
	GetUser(
		ctx context.Context,
		id int,
	) (user *models.User, err error)
}

type UserInfo struct {
	log              *slog.Logger
	userInfoProvider UserInfoProvider
	tokenTTL         time.Duration
}

func New(
	log *slog.Logger,
	userInfoProvider UserInfoProvider,
	tokenTTL time.Duration,
) *UserInfo {
	return &UserInfo{
		log:              log,
		userInfoProvider: userInfoProvider,
		tokenTTL:         tokenTTL,
	}
}

func (ui *UserInfo) GetUserInfo(ctx context.Context, token string) (*models.User, error) {
	const op = "UserInfo.GetUserInfo"

	log := ui.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)

	log.Info("decoding the jwt token")

	claims, err := auth.DecodeToken("sso-secret", token)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotValidJwt):
			return nil, auth.ErrNotValidJwt
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Info("decoded the jwt token")

	log.Info("retrieving user from storage")
	user, err := ui.userInfoProvider.GetUser(ctx, claims.UID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, sql.ErrNoRows
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Info("retrieved user from storage")
	return user, nil
}
