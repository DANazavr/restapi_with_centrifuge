package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/DANazavr/RATest/internal/common/meta"
	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/services"
	"github.com/centrifugal/gocent/v3"
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
	type response struct {
		Presence map[string]gocent.ClientInfo `json:"presence"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}
		body, err := json.Marshal(map[string]string{
			"channel": req.Channel,
		})
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to marshal request: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
			return
		}

		apiKey := os.Getenv("CENTRIFUGO_API_KEY")
		if apiKey == "" {
			nh.logger.Errorf(nh.ctx, "Centrifugo API key is not set")
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrInternalEnviroment)
			return
		}

		// Отправляем запрос в Centrifugo
		httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8000/api/presence", bytes.NewBuffer(body))
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to create request: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "apikey "+apiKey)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to call Centrifugo: %v", err)
			delivery.HendleError(w, r, http.StatusBadGateway, domain.ErrCentrifugePresenceFailed)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			nh.logger.Errorf(nh.ctx, "Centrifugo returned status: %d", resp.StatusCode)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrCentrifugePresenceFailed)
			return
		}

		var centrifugoResp response
		if err := json.NewDecoder(resp.Body).Decode(&centrifugoResp); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to decode Centrifugo response: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
			return
		}

		delivery.HendleRespond(w, r, http.StatusOK, centrifugoResp)
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
