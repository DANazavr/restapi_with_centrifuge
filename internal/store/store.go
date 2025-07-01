package store

type Store interface {
	User() UserRepository
	Notification() NotificationRepository
}
