package server

import (
	"context"
	"net/http"

	"github.com/DANazavr/RATest/config"
	"github.com/DANazavr/RATest/internal/delivery/http/auth"
	"github.com/DANazavr/RATest/internal/delivery/http/notification"
	"github.com/DANazavr/RATest/internal/delivery/http/user"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/internal/store"
	"github.com/gorilla/mux"
)

type server struct {
	ctx                 context.Context
	router              *mux.Router
	logger              *log.Log
	config              *config.Config
	authHendler         *auth.AuthHendler
	userHendler         *user.UserHendler
	notificationHandler *notification.NotificationHandler
	authMiddleware      *auth.MiddlewareAuth
	adminMiddleware     *auth.MiddlewareAdmin
}

func NewServer(ctx context.Context, store store.Store, config *config.Config, logger *log.Log, us *services.UserService, as *services.AuthService) *server {
	s := &server{
		ctx:                 ctx,
		router:              mux.NewRouter(),
		logger:              logger.WithComponent("http/server"),
		config:              config,
		authHendler:         auth.NewAuthHendler(ctx, logger, store, us, as),
		userHendler:         user.NewUserHendler(ctx, logger, store, us),
		notificationHandler: notification.NewNotificationHandler(ctx, logger, us, as.CentrifugeService),
		authMiddleware:      auth.NewMiddlewareAuth(ctx, logger, as),
		adminMiddleware:     auth.NewMiddlewareAdmin(ctx, logger, as),
	}

	s.configureRouter()

	s.logger.Infof(context.TODO(), "Server is running on port: %s", s.config.Addr)

	return s
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/register", s.authHendler.HandleRegister()).Methods("POST")
	s.router.HandleFunc("/login", s.authHendler.HandleLogin()).Methods("POST")
	s.router.HandleFunc("/token_refresh", s.authHendler.HandleTokensRefresh()).Methods("GET")

	in := s.router.PathPrefix("/in").Subrouter()
	in.Use(s.authMiddleware.Auth)
	in.HandleFunc("/profile", s.userHendler.HandleGetUser()).Methods("GET")

	admin := s.router.PathPrefix("/admin").Subrouter()
	admin.Use(s.adminMiddleware.Admin)
	admin.HandleFunc("/getUsers", s.userHendler.HandleGetUsers()).Methods("GET")

	notificationRouter := s.router.PathPrefix("/notification").Subrouter()
	notificationRouter.Use(s.adminMiddleware.Admin)
	// notificationRouter.HandleFunc("/broadcast", s.notificationHandler.SendBroadcastNotification()).Methods("POST")
	notificationRouter.HandleFunc("/presence", s.notificationHandler.Presence()).Methods("POST")
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
