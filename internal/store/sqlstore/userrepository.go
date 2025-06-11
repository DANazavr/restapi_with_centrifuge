package sqlstore

import (
	"github.com/DANazavr/RATest/internal/domain/models"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(user *models.User) error {
	if err := r.store.db.QueryRow(
		"INSERT INTO users (username, encrypted_password, email, role) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
		user.Username, user.EncryptedPassword, user.Email, user.Role,
	).Scan(&user.ID, &user.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	u := &models.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, username, encrypted_password, email, role, created_at FROM users WHERE username = $1", username,
	).Scan(
		&u.ID, &u.Username, &u.EncryptedPassword, &u.Email, &u.Role, &u.CreatedAt,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetById(id int) (*models.User, error) {
	u := &models.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, username, encrypted_password, email, role, created_at FROM users WHERE id = $1", id,
	).Scan(
		&u.ID, &u.Username, &u.EncryptedPassword, &u.Email, &u.Role, &u.CreatedAt,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Get() ([]*models.User, error) {
	u := make([]*models.User, 0, 100)
	rows, err := r.store.db.Query(
		"SELECT id, username, encrypted_password, email, role, created_at FROM users",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.EncryptedPassword, &user.Email, &user.Role, &user.CreatedAt,
		); err != nil {
			return nil, err
		}
		u = append(u, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return u, nil
}
