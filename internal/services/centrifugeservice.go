package services

import (
	"context"
	"encoding/json"

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

// func (cs *CentrifugeService) createNotificationChannel(userID ...int) ([]string, error) {
// 	channels := make([]string, 10, 100)
// 	for _, v := range userID {
// 		channels = append(channels, "notifications:user#"+strconv.Itoa(v))
// 	}
// 	cs.logger.Infof(cs.ctx, "Notification channel created: %s", channels)

// 	return channels, nil
// }

func (cs *CentrifugeService) Publish(n *models.UserNotification, channel string) (gocent.PublishResult, error) {
	data, err := json.Marshal(n.Notification)
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

func (cs *CentrifugeService) SaveNotification(n models.UserNotification) error {
	return nil
}
