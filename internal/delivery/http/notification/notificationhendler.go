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
	ctx                 context.Context
	logger              *log.Log
	userService         *services.UserService
	notificationService *services.NotificationService
}

func NewNotificationHandler(ctx context.Context, logger *log.Log, us *services.UserService, cs *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		ctx:                 ctx,
		logger:              logger.WithComponent("rest/notification/notificationHandler"),
		userService:         us,
		notificationService: cs,
	}
}

func (nh *NotificationHandler) Publish() http.HandlerFunc {
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

		if _, err := nh.userService.UsersGetById(userID); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get user: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrUserNotFound)
			return
		}
		ctx = context.WithValue(ctx, meta.UserIDKey, userID)

		notificationMap := map[string]interface{}{
			"title":   req.Data.Title,
			"message": req.Data.Message,
		}
		n := &models.UserNotification{
			UserID:       userID,
			Notification: notificationMap,
		}

		if err := nh.notificationService.NotificationCreate(n); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to create notification: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotificationCreateFailed)
			return
		}
		nh.logger.Infof(nh.ctx, "Notification created for user %d: %v", userID, n.UID)

		presence, err := nh.notificationService.Presence(req.Channel)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get presence: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
			return
		}
		if len(presence.Presence) == 0 {
			nh.logger.Infof(nh.ctx, "No users online in channel %s", req.Channel)
			delivery.HendleError(w, r, http.StatusNotFound, domain.ErrCentrifugePresenceFailed)
			return
		}

		publish, err := nh.notificationService.Publish(n, req.Channel)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to publish notification: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrCentrifugePublishFailed)
			return
		}

		for _, v := range presence.Presence {
			userid, err := strconv.Atoi(v.User)
			if err != nil {
				nh.logger.Errorf(nh.ctx, "Failed to convert user ID from presence: %v", err)
				delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
				return
			}
			if err := nh.notificationService.MarkAsSend(n, userid); err != nil {
				nh.logger.Errorf(nh.ctx, "Failed to mark notification as sent: %v", err)
				delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(publish); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to encode response: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePublishFailed)
			return
		}

		r = r.WithContext(ctx)
		nh.logger.Infof(r.Context(), "Notification sent to user %d: %v", userID, n.Notification)
	}
}

func (nh *NotificationHandler) Broadcast() http.HandlerFunc {
	type notification struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	type request struct {
		Data notification `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}

		users, err := nh.userService.UsersGet()
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get users: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrUserNotFound)
			return
		}
		for _, user := range users {
			if user.Role != "user" {
				nh.logger.Infof(nh.ctx, "Skipping user %d with role %s", user.ID, user.Role)
				continue
			}
			notificationMap := map[string]interface{}{
				"title":   req.Data.Title,
				"message": req.Data.Message,
			}
			n := &models.UserNotification{
				UserID:       user.ID,
				Notification: notificationMap,
			}

			channel := "notifications:user#" + strconv.Itoa(user.ID)

			if err := nh.notificationService.NotificationCreate(n); err != nil {
				nh.logger.Errorf(nh.ctx, "Failed to create notification: %v", err)
				delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotificationCreateFailed)
				return
			}
			nh.logger.Infof(nh.ctx, "Notification created for user %d: %v", user.ID, n.UID)
			_, err := nh.notificationService.Publish(n, channel)
			if err != nil {
				nh.logger.Errorf(nh.ctx, "Failed to publish notification: %v", err)
				delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrCentrifugePublishFailed)
				return
			}
			presence, err := nh.notificationService.Presence(channel)
			if err != nil {
				nh.logger.Errorf(nh.ctx, "Failed to get presence: %v", err)
				delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePresenceFailed)
				return
			}
			if len(presence.Presence) == 0 {
				nh.logger.Infof(nh.ctx, "No users online in channel %s", channel)
				delivery.HendleError(w, r, http.StatusNotFound, domain.ErrCentrifugePresenceFailed)
				return
			}
			for _, v := range presence.Presence {
				userid, err := strconv.Atoi(v.User)
				if err != nil {
					nh.logger.Errorf(nh.ctx, "Failed to convert user ID from presence: %v", err)
					delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
					return
				}
				if err := nh.notificationService.MarkAsSend(n, userid); err != nil {
					nh.logger.Errorf(nh.ctx, "Failed to mark notification as sent: %v", err)
					delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")

		r = r.WithContext(ctx)
		nh.logger.Infof(r.Context(), "Broadcast notification sent: %v", req.Data)
	}
}

func (nh *NotificationHandler) MarkAsRead() http.HandlerFunc {
	type request struct {
		NotificationID int `json:"notification_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userIDstr, ok := ctx.Value(meta.UserIDKey).(string)
		if !ok || userIDstr == "" {
			nh.logger.Errorf(nh.ctx, "Invalid user ID in context: %v", userIDstr)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidUserID)
			return
		}
		userID, err := strconv.Atoi(userIDstr)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to convert user ID to int: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidUserID)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}
		_, err = nh.userService.UsersGetById(userID)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get user: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrUserNotFound)
			return
		}
		notification, err := nh.notificationService.GetById(req.NotificationID)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get notifications: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}

		if err := nh.notificationService.MarkAsRead(notification, userID); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to mark notification as read: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(notification); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to encode response: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}
	}
}

func (nh *NotificationHandler) GetNotificationsByFilter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userIDstr, ok := ctx.Value(meta.UserIDKey).(string)
		if !ok || userIDstr == "" {
			nh.logger.Errorf(nh.ctx, "Invalid user ID in context: %v", userIDstr)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidUserID)
			return
		}
		userID, err := strconv.Atoi(userIDstr)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to convert user ID to int: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidUserID)
			return
		}

		_, err = nh.userService.UsersGetById(userID)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get user: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrUserNotFound)
			return
		}

		params := r.URL.Query()
		filter := params.Get("filter")
		if !nh.notificationService.ValidateFilter(filter) {
			nh.logger.Errorf(nh.ctx, "Invalid filter: %s", filter)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidFilter)
			return
		}

		notifications, err := nh.notificationService.GetByUserIdWithFilter(userID, filter)
		if err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to get notifications: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}

		for _, v := range notifications {
			if v.SendAt == nil {
				if err := nh.notificationService.MarkAsSend(v, userID); err != nil {
					nh.logger.Errorf(nh.ctx, "Failed to mark notification as sent: %v", err)
					delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(notifications); err != nil {
			nh.logger.Errorf(nh.ctx, "Failed to encode response: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}
	}
}
