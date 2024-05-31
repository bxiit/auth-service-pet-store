package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/data/storage"
	"sso/internal/services/auth"
	"sso/internal/services/user_info"
	"time"
)

// App wrapper for grpcapp.App
type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
) *App {
	// TODO: database setup
	authStorage, err := storage.NewAuthStorage(dsn)
	if err != nil {
		panic(err)
	}

	userInfoStorage, err := storage.NewUserInfoStorage(dsn)
	if err != nil {
		panic(err)
	}

	// TODO: auth service setup
	authService := auth.New(log, tokenTTL, authStorage)

	userInfoService := user_info.New(log, userInfoStorage, tokenTTL)

	// TODO: grpc app setup
	grpcApp := grpcapp.New(log, authService, userInfoService, grpcPort)

	go authStorage.CheckTokens()
	return &App{
		GRPCServer: grpcApp,
	}
}
