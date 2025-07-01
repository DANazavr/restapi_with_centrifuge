package sqlstore

import (
	"encoding/json"

	"github.com/DANazavr/RATest/internal/domain/models"
)

type NotificationRepository struct {
	store *Store
}

func (n *NotificationRepository) Create(un *models.UserNotification, data []byte) error {
	err := n.store.db.QueryRow(
		"INSERT INTO user_notifications (user_id, notification) VALUES ($1, $2) RETURNING uid, created_at",
		&un.UserID, &data,
	).Scan(&un.UID, &un.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (n *NotificationRepository) GetById(id int) (*models.UserNotification, error) {
	var data []byte
	un := &models.UserNotification{}
	err := n.store.db.QueryRow(
		"SELECT uid, user_id, notification, created_at, send_at, read_at FROM user_notifications WHERE uid = $1", id,
	).Scan(&un.UID, &un.UserID, &data, &un.CreatedAt, &un.SendAt, &un.ReadAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &un.Notification); err != nil {
		return nil, err
	}
	return un, nil
}

func (n *NotificationRepository) GetByUserId(userId int) ([]*models.UserNotification, error) {
	un := make([]*models.UserNotification, 0, 100)
	rows, err := n.store.db.Query(
		"SELECT uid, user_id, notification, created_at, send_at, read_at FROM user_notifications WHERE user_id = $1 ORDER BY created_at", userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var data []byte
		userNotification := &models.UserNotification{}
		if err := rows.Scan(
			&userNotification.UID, &userNotification.UserID, &data,
			&userNotification.CreatedAt, &userNotification.SendAt, &userNotification.ReadAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &userNotification.Notification); err != nil {
			return nil, err
		}
		un = append(un, userNotification)
	}
	return un, nil
}

func (n *NotificationRepository) GetByUserIdWithFilter(userId int, filter string) ([]*models.UserNotification, error) {
	un := make([]*models.UserNotification, 0, 100)
	query := "SELECT uid, user_id, notification, created_at, send_at, read_at FROM user_notifications WHERE user_id = $1"
	switch filter {
	case "all":
		// No additional conditions
	case "unread":
		query += " AND read_at IS NULL"
	case "read":
		query += " AND read_at IS NOT NULL"
	case "unsend":
		query += " AND send_at IS NULL"
	case "send":
		query += " AND send_at IS NOT NULL"
	case "sendandread":
		query += " AND send_at IS NOT NULL AND read_at IS NOT NULL"
	case "sendandunread":
		query += " AND send_at IS NOT NULL AND read_at IS NULL"
	case "unsendandread":
		query += " AND send_at IS NULL AND read_at IS NOT NULL"
	case "unsendandunread":
		query += " AND send_at IS NULL AND read_at IS NULL"
	default:
		return nil, nil // Invalid filter
	}

	query += " ORDER BY created_at"

	rows, err := n.store.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var data []byte
		userNotification := &models.UserNotification{}
		if err := rows.Scan(
			&userNotification.UID, &userNotification.UserID, &data,
			&userNotification.CreatedAt, &userNotification.SendAt, &userNotification.ReadAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &userNotification.Notification); err != nil {
			return nil, err
		}
		un = append(un, userNotification)
	}
	return un, nil
}

func (n *NotificationRepository) MarkAsSend(id int, userid int) error {
	_, err := n.store.db.Exec(
		"UPDATE user_notifications SET send_at = NOW() WHERE uid = $1 AND user_id = $2", id, userid,
	)
	if err != nil {
		return err
	}
	return nil
}

func (n *NotificationRepository) MarkAsRead(id int, userid int) error {
	_, err := n.store.db.Exec(
		"UPDATE user_notifications SET read_at = NOW() WHERE uid = $1 AND user_id = $2", id, userid,
	)
	if err != nil {
		return err
	}
	return nil
}
