package server

import (
	"context"

	"github.com/DANazavr/RATest/config"
	"github.com/DANazavr/RATest/internal/delivery/grpc/server/admin"
	"github.com/DANazavr/RATest/internal/delivery/grpc/server/auth"
	"github.com/DANazavr/RATest/internal/delivery/grpc/server/notification"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"google.golang.org/grpc"
)

type Server struct {
	ctx                 context.Context
	logger              *log.Log
	config              *config.Config
	adminInterceptor    *admin.InterceptorAdmin
	authInterceptor     *auth.InterceptorAuth
	authHendler         *auth.AuthServer
	notificationHendler *notification.NotificationServer
	gRPCServer          *grpc.Server
}

func NewServer(ctx context.Context, logger *log.Log, config *config.Config, as *services.AuthService, us *services.UserService, ns *services.NotificationService) *Server {
	s := &Server{
		ctx:                 ctx,
		logger:              logger.WithComponent("grpc/server/Server"),
		config:              config,
		adminInterceptor:    admin.NewInterceptorAdmin(ctx, logger, as),
		authInterceptor:     auth.NewInterceptorAuth(ctx, logger, as),
		authHendler:         auth.NewAuthServer(ctx, logger, us, as),
		notificationHendler: notification.NewNotificationServer(ctx, logger, us, ns),
	}

	s.gRPCServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.adminInterceptor.AdminInterceptor,
			s.authInterceptor.AuthInterceptor,
		),
	)

	auth.Register(s.gRPCServer, s.authHendler)
	notification.Register(s.gRPCServer, s.notificationHendler)

	return s
}

func (s *Server) GetGRPCServer() *grpc.Server {
	return s.gRPCServer
}
