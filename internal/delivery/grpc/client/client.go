package client

import (
	"context"
	"net/http"

	"github.com/DANazavr/RATest/config"
	"github.com/DANazavr/RATest/internal/delivery/grpc/client/auth"
	"github.com/DANazavr/RATest/internal/delivery/grpc/client/notification"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type client struct {
	ctx                context.Context
	logger             *log.Log
	config             *config.Config
	handler            http.Handler
	router             *mux.Router
	authClient         *auth.AuthClient
	notificationClient *notification.NotificationClient
}

func NewAuthClient(ctx context.Context, logger *log.Log, config *config.Config) *client {
	authClient, err := auth.NewAuthClient(ctx, logger)
	if err != nil {
		return nil
	}
	notificationClient, err := notification.NewNotificationClient(ctx, logger)
	if err != nil {
		return nil
	}

	c := &client{
		ctx:                ctx,
		logger:             logger.WithComponent("grpc/client/auth"),
		config:             config,
		router:             mux.NewRouter(),
		authClient:         authClient,
		notificationClient: notificationClient,
	}

	c.configureRouter()

	return c
}

func (c *client) configureRouter() {
	c.router.HandleFunc("/register", c.authClient.Register()).Methods("POST")
	c.router.HandleFunc("/login", c.authClient.Login()).Methods("POST")
	c.router.HandleFunc("/token_refresh", c.authClient.TokenRefresh()).Methods("GET")

	in := c.router.PathPrefix("/user").Subrouter()
	in.Use(auth.AuthMiddleware)
	in.HandleFunc("/getnotifications", c.notificationClient.GetNotificationsByFilter()).Methods("GET")
	in.HandleFunc("/markasread", c.notificationClient.MarkAsRead()).Methods("POST")
	// in.HandleFunc("/profile", c.userHendler.HandleGetUser()).Methods("GET")

	// admin := c.router.PathPrefix("/admin").Subrouter()
	// admin.HandleFunc("/getUsers", c.userHendler.HandleGetUsers()).Methods("GET")

	notificationRouter := c.router.PathPrefix("/notification").Subrouter()
	notificationRouter.Use(auth.AuthMiddleware)
	notificationRouter.HandleFunc("/broadcast", c.notificationClient.Broadcast()).Methods("POST")
	notificationRouter.HandleFunc("/publish", c.notificationClient.Publish()).Methods("POST")

	co := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	c.handler = co.Handler(c.router)
}

func (c *client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.handler.ServeHTTP(w, r)
}
