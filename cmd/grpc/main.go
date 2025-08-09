package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/DANazavr/RATest/config"
	grpcapp "github.com/DANazavr/RATest/internal/app/grpc"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/internal/store/sqlstore"
	"github.com/joho/godotenv"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "./config/config.json", "path to config file")
}

func main() {
	flag.Parse()
	config := config.ParseConfig(configPath)

	ctx, cansel := context.WithCancel(context.Background())
	logger := log.NewLog(ctx, &log.LogConfig{Component: "main"})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		<-quit
		cansel()
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Keep the goroutine alive to handle signals
			}
		}
	}()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal(ctx, "Error loading .env file")
	}

	db, err := newDB(config.DatabaseURL)
	if err != nil {
		logger.Fatalf(ctx, "Failed to connect to database: %v", err)
	}
	defer db.Close()
	store := sqlstore.New(ctx, db, logger)

	userService := services.NewUserService(ctx, store, logger)
	notificationService := services.NewNotificationService(ctx, logger, store)
	authService := services.NewAuthService(ctx, logger)

	go func() {
		if err := grpcapp.Start(ctx, logger, config, store, userService, authService, notificationService); err != nil {
			logger.Fatalf(ctx, "Failed to start server: %v", err)
		}
	}()
	wg.Wait()
	defer logger.Info(ctx, "Server stopped gracefully")
}

func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
