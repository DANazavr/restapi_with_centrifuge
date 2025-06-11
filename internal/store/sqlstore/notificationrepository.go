package sqlstore

import (
	"github.com/DANazavr/RATest/internal/domain/models"
)

type NotificationRepository struct {
	store *Store
}

func (n *NotificationRepository) Create(un *models.UserNotification) error {
	err := n.store.db.QueryRow(
		"INSERT INTO user_notifications (user_id, notification) VALUES ($1, $2, $3) RETURNING id, created_at",
		&un.UserID, &un.Notification,
	).Scan(&un.UID, &un.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (n *NotificationRepository) GetById(id int) (*models.UserNotification, error) {
	un := &models.UserNotification{}
	err := n.store.db.QueryRow(
		"SELECT uid, user_id, notification, created_at, send_at, read_at FROM user_notifications WHERE uid = $1", id,
	).Scan(&un.UID, &un.UserID, &un.Notification, &un.CreatedAt, &un.SendAt, &un.ReadAt)
	if err != nil {
		return nil, err
	}
	return un, nil
}

func (n *NotificationRepository) GetByUserId(userId int) ([]*models.UserNotification, error) {
	un := make([]*models.UserNotification, 0, 100)
	rows, err := n.store.db.Query(
		"SELECT uid, user_id, notification, created_at, send_at, read_at FROM user_notifications WHERE user_id = $1", userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userNotification := &models.UserNotification{}
		if err := rows.Scan(
			&userNotification.UID, &userNotification.UserID, &userNotification.Notification,
			&userNotification.CreatedAt, &userNotification.SendAt, &userNotification.ReadAt,
		); err != nil {
			return nil, err
		}
		un = append(un, userNotification)
	}
	return un, nil
}

func (n *NotificationRepository) MarkAsSend(id int) error {
	_, err := n.store.db.Exec(
		"UPDATE user_notifications SET send_at = NOW() WHERE uid = $1", id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (n *NotificationRepository) MarkAsRead(id int) error {
	_, err := n.store.db.Exec(
		"UPDATE user_notifications SET read_at = NOW() WHERE uid = $1", id,
	)
	if err != nil {
		return err
	}
	return nil
}
