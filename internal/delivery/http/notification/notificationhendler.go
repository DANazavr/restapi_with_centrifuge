package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DANazavr/RATest/internal/common/meta"
	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
)

type NotificationHandler struct {
	ctx               context.Context
	logger            *log.Log
	userService       *services.UserService
	centrifugeService *services.CentrifugeService
}

func NewNotificationHandler(ctx context.Context, logger *log.Log, us *services.UserService, cs *services.CentrifugeService) *NotificationHandler {
	return &NotificationHandler{
		ctx:               ctx,
		logger:            logger.WithComponent("notification/notificationHandler"),
		userService:       us,
		centrifugeService: cs,
	}
}

func (nh *NotificationHandler) Presence() http.HandlerFunc {
	type request struct {
		Channel string `json:"channel"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}

		if req.Channel == "" {
			nh.logger.Errorf(nh.ctx, "Channel is empty")
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}

		presence, err := nh.centrifugeService.Presence(req.Channel)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get presence: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		for _, v := range presence.Presence {
			nh.logger.Info(nh.ctx, "user: ", v.User)
			if err := json.NewEncoder(w).Encode(v.User); err != nil {
				nh.logger.Errorf(nh.ctx, "Failed to encode response: %v", err)
				delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
				return
			}
		}
	}
}

func (nh *NotificationHandler) Notify() http.HandlerFunc {
	type notification struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	type request struct {
		Channel string       `json:"channel"`
		Data    notification `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}

		var userID int
		for i, v := range req.Channel {
			if v == '#' {
				var err error
				userID, err = strconv.Atoi(req.Channel[i+1:])
				if err != nil {
					nh.logger.Errorf(nh.ctx, "Failed to decode userid: %v", err)
					delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
				}
			}
		}

		if userID <= 0 {
			nh.logger.Errorf(nh.ctx, "Invalid user ID: %d", userID)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidUserID)
			return
		}

		user, err := nh.userService.UsersGetById(userID)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get user: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrUserNotFound)
			return
		}
		ctx = context.WithValue(ctx, meta.UserIDKey, user.ID)

		notificationMap := map[string]interface{}{
			"title":   req.Data.Title,
			"message": req.Data.Message,
		}
		n := &models.UserNotification{
			UserID:       user.ID,
			Notification: notificationMap,
		}

		if err := nh.centrifugeService.PublishNotification(n); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to publish notification: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrCentrifugePublishFailed)
			return
		}

		nh.logger.Infof(ctx, "Notification sent to user %d: %v", user.ID, n.Notification)
		w.WriteHeader(http.StatusOK)
	}
}
