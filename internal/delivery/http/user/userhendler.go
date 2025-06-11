package user

import (
	"context"
	"net/http"

	"github.com/DANazavr/RATest/internal/common/meta"
	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/DANazavr/RATest/internal/store"
)

type UserHendler struct {
	ctx         context.Context
	logger      *log.Log
	userService *services.UserService
}

func NewUserHendler(ctx context.Context, logger *log.Log, store store.Store, us *services.UserService) *UserHendler {
	return &UserHendler{
		ctx:         ctx,
		logger:      logger.WithComponent("user/userHendler"),
		userService: us,
	}
}

func (h *UserHendler) HandleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, ok := ctx.Value(meta.UserIDKey).(int64)
		if !ok {
			delivery.HendleError(w, r, http.StatusUnauthorized, domain.ErrInvalidToken)
			return
		}
		u, err := h.userService.UsersGetById(int(userID))
		if err != nil {
			delivery.HendleError(w, r, http.StatusNotFound, err)
			return
		}
		delivery.HendleRespond(w, r, http.StatusOK, u)
	}
}

func (h *UserHendler) HandleGetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.userService.UsersGet()
		if err != nil {
			delivery.HendleError(w, r, http.StatusNotFound, err)
			return
		}
		delivery.HendleRespond(w, r, http.StatusOK, u)
	}
}
