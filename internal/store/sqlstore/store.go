package sqlstore

import (
	"context"
	"database/sql"

	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/store"
	_ "github.com/lib/pq"
)

type Store struct {
	ctx                    context.Context
	logger                 *log.Log
	db                     *sql.DB
	userRepository         *UserRepository
	notificationRepository *NotificationRepository
}

func New(ctx context.Context, db *sql.DB, logger *log.Log) *Store {
	return &Store{
		ctx:    ctx,
		db:     db,
		logger: logger.WithComponent("sqlstore"),
	}
}

func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}
	s.userRepository = &UserRepository{
		store: s,
	}
	return s.userRepository
}

func (s *Store) Notification() store.NotificationRepository {
	if s.notificationRepository != nil {
		return s.notificationRepository
	}
	s.notificationRepository = &NotificationRepository{
		store: s,
	}
	return s.notificationRepository
}
