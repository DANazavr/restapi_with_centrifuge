package rest

import (
	"context"
	"net/http"

	"github.com/DANazavr/RATest/config"
	"github.com/DANazavr/RATest/internal/delivery/http/server"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/internal/store"
)

func Start(ctx context.Context, store store.Store, config *config.Config, logger *log.Log, us *services.UserService, as *services.AuthService) error {
	srv := server.NewServer(ctx, store, config, logger, us, as)
	return http.ListenAndServe(config.Addr, srv)
}
