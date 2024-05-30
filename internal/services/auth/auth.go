package auth

// auth implementation

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/data/models"
	"sso/internal/data/storage"
	"sso/internal/sl"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Auth service itself
//type Auth struct {
//	log           *slog.Logger
//	usrSaver      UserSaver
//	usrProvider   UserProvider
//	appProvider   AppProvider
//	tokenProvider TokenProvider
//	tokenSaver    TokenSaver
//	tokenTTL      time.Duration
//}

type Auth struct {
	log         *slog.Logger
	ssoProvider SsoProvider
	tokenTTL    time.Duration
}

type SsoProvider interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	App(ctx context.Context, appID int) (models.App, error)
	SaveToken(ctx context.Context, tokenPlainText string, userId int64) (bool, error)
	IsAuthenticated(ctx context.Context, token string) (bool, error)
}

// New returns Auth service
//func New(
//	log *slog.Logger,
//	userSaver UserSaver,
//	userProvider UserProvider,
//	appProvider AppProvider,
//	tokenProvider TokenProvider,
//	tokenSaver TokenSaver,
//	tokenTTL time.Duration,
//) *Auth {
//	return &Auth{
//		usrSaver:      userSaver,
//		usrProvider:   userProvider,
//		log:           log,
//		appProvider:   appProvider,
//		tokenTTL:      tokenTTL,
//		tokenProvider: tokenProvider,
//		tokenSaver:    tokenSaver,
//	}
//}

func New(
	log *slog.Logger,
	tokenTTL time.Duration,
	ssoProvider SsoProvider,
) *Auth {
	return &Auth{
		log:         log,
		tokenTTL:    tokenTTL,
		ssoProvider: ssoProvider,
	}
}

// UserSaver interface for service methods
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
//type UserSaver interface {
//	SaveUser(
//		ctx context.Context,
//		email string,
//		passHash []byte,
//	) (uid int64, err error)
//}
//
//// UserProvider interface for service methods
//type UserProvider interface {
//	GetUserByEmail(ctx context.Context, email string) (models.User, error)
//	IsAdmin(ctx context.Context, userID int64) (bool, error)
//}
//
//type TokenProvider interface {
//	IsAuthenticated(ctx context.Context, token string) (bool, error)
//}
//
//type TokenSaver interface {
//	SaveToken(ctx context.Context, tokenPlainText string, userId int64) (bool, error)
//}
//
//// AppProvider interface for service methods
//type AppProvider interface {
//	App(ctx context.Context, appID int) (models.App, error)
//}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.ssoProvider.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash.Hash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.ssoProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	isSaved, err := a.ssoProvider.SaveToken(ctx, token, user.ID)
	if err != nil || !isSaved {
		a.log.Warn("token not saved", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.ssoProvider.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.ssoProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) IsAuthenticated(ctx context.Context, token string) (bool, error) {
	const op = "Auth.IsAuthenticated"

	log := a.log.With(
		slog.String("op", op),
		slog.String("token", token),
	)

	log.Info("checking if user is authenticated")

	isAdmin, err := a.ssoProvider.IsAuthenticated(ctx, token)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is authenticated", slog.Bool("is_authenticated", isAdmin))

	return isAdmin, nil
}
