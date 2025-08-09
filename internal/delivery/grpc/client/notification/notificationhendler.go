package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	delivery "github.com/DANazavr/RATest/internal/delivery/http"
	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/protos/gen/go/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	ctx    context.Context
	logger *log.Log
	client notification.NotificationClient
}

func NewNotificationClient(ctx context.Context, logger *log.Log) (*NotificationClient, error) {
	conn, err := grpc.NewClient(":8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &NotificationClient{
		ctx:    ctx,
		logger: logger.WithComponent("grpc/client/notification"),
		client: notification.NewNotificationClient(conn),
	}, nil
}

func (nc *NotificationClient) Publish() http.HandlerFunc {
	type notif struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	type request struct {
		Channel string `json:"channel"`
		Data    notif  `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nc.logger.Errorf(nc.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}

		resp, err := nc.client.Publish(ctx, &notification.PublishRequest{
			Channel: req.Channel,
			Data: &notification.Data{
				Title:   req.Data.Title,
				Message: req.Data.Message,
			},
		})
		if err != nil {
			nc.logger.Errorf(ctx, "Failed to publish notification: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePublishFailed)
			return
		}
		delivery.HendleRespond(w, r, http.StatusCreated, resp)
	}
}

func (nc *NotificationClient) Broadcast() http.HandlerFunc {
	type notif struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	type request struct {
		Data notif `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nc.logger.Errorf(nc.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}

		resp, err := nc.client.Broadcast(ctx, &notification.BroadcastRequest{
			Data: &notification.Data{
				Title:   req.Data.Title,
				Message: req.Data.Message,
			},
		})
		if err != nil {
			nc.logger.Errorf(ctx, "Failed to broadcast notification: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugePublishFailed)
			return
		}
		delivery.HendleRespond(w, r, http.StatusCreated, resp)
	}
}

func (nc *NotificationClient) MarkAsRead() http.HandlerFunc {
	type request struct {
		NotificationID int `json:"notification_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			nc.logger.Errorf(nc.ctx, "Failed to decode request: %v", err)
			delivery.HendleError(w, r, http.StatusBadRequest, domain.ErrInvalidRequestBody)
			return
		}
		resp, err := nc.client.MarkAsRead(ctx, &notification.MarkAsReadRequest{
			NotificationId: fmt.Sprintf("%d", req.NotificationID),
		})
		if err != nil {
			nc.logger.Errorf(ctx, "Failed to mark notification as read: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}
		delivery.HendleRespond(w, r, http.StatusCreated, resp)
	}
}

func (nc *NotificationClient) GetNotificationsByFilter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := r.URL.Query()
		filter := params.Get("filter")

		resp, err := nc.client.GetNotificationsByFilter(ctx, &notification.GetNotificationsByFilterRequest{
			Filter: filter,
		})
		if err != nil {
			nc.logger.Errorf(ctx, "Failed to get notifications by filter: %v", err)
			delivery.HendleError(w, r, http.StatusInternalServerError, domain.ErrCentrifugeNotification)
			return
		}
		// notifications :=
		delivery.HendleRespond(w, r, http.StatusCreated, resp.Notifications)
	}
}
