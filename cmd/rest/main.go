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
	"github.com/DANazavr/RATest/internal/app/rest"
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
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
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
	centrifugeService := services.NewCentrifugeService(ctx, logger, store)
	authService := services.NewAuthService(ctx, logger, centrifugeService)

	go func() {
		if err := rest.Start(ctx, store, config, logger, userService, authService); err != nil {
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
