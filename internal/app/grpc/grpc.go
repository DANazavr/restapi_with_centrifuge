package grpcapp

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/DANazavr/RATest/config"
	"github.com/DANazavr/RATest/internal/delivery/grpc/client"
	"github.com/DANazavr/RATest/internal/delivery/grpc/server"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/internal/store"
)

func Start(ctx context.Context, logger *log.Log, config *config.Config, store store.Store, us *services.UserService, as *services.AuthService, ns *services.NotificationService) error {
	grpcServer := server.NewServer(ctx, logger, config, as, us, ns)
	grpcListener, err := net.Listen("tcp", config.GRPCAddr)
	if err != nil {
		logger.Fatalf(ctx, "Failed to listen on %s: %v", config.GRPCAddr, err)
		return err
	}

	httpServer := client.NewAuthClient(ctx, logger, config)

	// Канал для ошибок
	errChan := make(chan error, 2)

	// Запуск gRPC сервера в горутине
	go func() {
		logger.Infof(ctx, "gRPC server starting on %s", config.GRPCAddr)
		if err := grpcServer.GetGRPCServer().Serve(grpcListener); err != nil {
			errChan <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// Запуск HTTP сервера в горутине
	go func() {
		logger.Infof(ctx, "HTTP server starting on %s", config.RestAddr)
		if err := http.ListenAndServe(config.RestAddr, httpServer); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Ожидание завершения
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// Graceful shutdown
		logger.Info(ctx, "Shutting down servers...")

		// Остановка gRPC сервера
		grpcServer.GetGRPCServer().GracefulStop()

		return nil
	}
}
