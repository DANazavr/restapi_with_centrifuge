package services

import (
	"context"
	"encoding/json"

	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/store"
	"github.com/centrifugal/gocent/v3"
)

type CentrifugeService struct {
	ctx    context.Context
	logger *log.Log
	store  store.Store
	Client *gocent.Client
}

func NewCentrifugeService(ctx context.Context, logger *log.Log, store store.Store) *CentrifugeService {
	c := gocent.New(gocent.Config{
		Addr: "http://localhost:8000/api", // правильный адрес для gocent клиента
		Key:  "my_api_key",
	})

	return &CentrifugeService{
		ctx:    ctx,
		logger: logger.WithComponent("services/centrifuge"),
		store:  store,
		Client: c,
	}
}

func (cs *CentrifugeService) Presence(channel string) (gocent.PresenceResult, error) {
	presence, err := cs.Client.Presence(cs.ctx, channel)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to get presence into %v: %v", channel, err)
		return gocent.PresenceResult{}, err
	}
	return presence, nil
}

func (cs *CentrifugeService) Publish(n *models.UserNotification, channel string) (gocent.PublishResult, error) {
	data, err := json.Marshal(n)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to marshal notification: %v", err)
		return gocent.PublishResult{}, err
	}

	publish, err := cs.Client.Publish(cs.ctx, channel, data)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to publish notification to channel %s: %v", channel, err)
		return gocent.PublishResult{}, err
	}

	return publish, nil
}

func (cs *CentrifugeService) NotificationCreate(n *models.UserNotification) error {
	data, err := json.Marshal(n.Notification)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to marshal notification: %v", err)
		return err
	}
	if err := cs.store.Notification().Create(n, data); err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to create notification: %v", err)
		return err
	}
	return nil
}

func (cs *CentrifugeService) GetByUserId(userID int) ([]*models.UserNotification, error) {
	n, err := cs.store.Notification().GetByUserId(userID)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to get notifications for user %d: %v", userID, err)
		return nil, err
	}
	return n, nil
}

func (cs *CentrifugeService) GetByUserIdWithFilter(userID int, filter string) ([]*models.UserNotification, error) {
	n, err := cs.store.Notification().GetByUserIdWithFilter(userID, filter)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to get notifications for user %d: %v", userID, err)
		return nil, err
	}
	return n, nil
}

func (cs *CentrifugeService) GetById(id int) (*models.UserNotification, error) {
	n, err := cs.store.Notification().GetById(id)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to get notification by ID %d: %v", id, err)
		return nil, err
	}
	return n, nil
}

func (cs *CentrifugeService) MarkAsSend(n *models.UserNotification, userID int) error {
	if err := cs.store.Notification().MarkAsSend(n.UID, userID); err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to mark notification as sent: %v", err)
		return err
	}
	return nil
}

func (cs *CentrifugeService) MarkAsRead(n *models.UserNotification, userID int) error {
	if err := cs.store.Notification().MarkAsRead(n.UID, userID); err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to mark notification as sent: %v", err)
		return err
	}
	return nil
}

func (cs *CentrifugeService) ValidateFilter(filter string) bool {
	validFilters := []string{"all", "unread", "read", "unsend", "send", "sendandread", "sendandunread", "unsendandread", "unsendandunread"}
	for _, f := range validFilters {
		if f == filter {
			return true
		}
	}
	return false
}
