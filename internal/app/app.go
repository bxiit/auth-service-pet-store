package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/data/storage"
	"sso/internal/services/auth"
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
	storage, err := storage.New(dsn)
	if err != nil {
		panic(err)
	}

	// TODO: auth service setup
	authService := auth.New(log, tokenTTL, storage)

	// TODO: grpc app setup
	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
