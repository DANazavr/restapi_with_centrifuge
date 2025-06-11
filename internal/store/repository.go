package store

import "github.com/DANazavr/RATest/internal/domain/models"

type UserRepository interface {
	Create(*models.User) error
	GetByUsername(string) (*models.User, error)
	GetById(int) (*models.User, error)
	Get() ([]*models.User, error)
}

type NotificationRepository interface {
	Create(*models.UserNotification) error
	GetById(int) (*models.UserNotification, error)
	GetByUserId(int) ([]*models.UserNotification, error)
	MarkAsSend(int) error
	MarkAsRead(int) error
}
