package services

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/centrifugal/gocent/v3"
)

type CentrifugeService struct {
	ctx    context.Context
	logger *log.Log
	Client *gocent.Client
}

func NewCentrifugeService(ctx context.Context, logger *log.Log) *CentrifugeService {
	c := gocent.New(gocent.Config{
		Addr: "http://localhost:8000/api", // правильный адрес для gocent клиента
		Key:  "my_api_key",
	})

	return &CentrifugeService{
		ctx:    ctx,
		logger: logger.WithComponent("services/centrifuge"),
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

func (cs *CentrifugeService) createNotificationChannel(userID ...int) ([]string, error) {
	channels := make([]string, 10, 100)
	for _, v := range userID {
		channels = append(channels, "notifications:user#"+strconv.Itoa(v))
	}
	cs.logger.Infof(cs.ctx, "Notification channel created: %s", channels)

	return channels, nil
}

func (cs *CentrifugeService) PublishNotification(n *models.UserNotification) error {
	channel, err := cs.createNotificationChannel(n.UserID)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to create notification channel: %v", err)
		return err
	}
	data, err := json.Marshal(n)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to marshal notification: %v", err)
		return err
	}
	channels, err := cs.Client.Channels(cs.ctx)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to get channels: %v", err)
		return err
	}
	var channelNames []string
	for ch := range channels.Channels {
		channelNames = append(channelNames, ch)
	}
	broadcastResult, err := cs.Client.Broadcast(cs.ctx, channelNames, data)
	if err != nil {
		cs.logger.Errorf(cs.ctx, "Failed to broadcast notification to channel %s: %v", channel, err)
		return err
	}
	cs.logger.Infof(cs.ctx, "Broadcast result into this channels %s: %v", channel, broadcastResult)
	cs.logger.Infof(cs.ctx, "Notification published to channel %s: %s", channel, string(data))
	return nil
}
