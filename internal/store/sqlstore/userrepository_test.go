package sqlstore_test

import (
	"testing"

	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/store/sqlstore"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
	t.Helper()

	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("users")

	store := sqlstore.New(t.Context(), db, log.NewLog(t.Context(), &log.LogConfig{Component: "sqlstore", LogLevel: "debug"}))
	u := &models.User{
		Username:          "testuser",
		EncryptedPassword: "encrypted_password",
		Email:             "email@example.com",
		Role:              "user",
	}
	// Create a new user
	err := store.User().Create(u)

	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("users")

	s := sqlstore.New(t.Context(), db, log.NewLog(t.Context(), &log.LogConfig{Component: "sqlstore", LogLevel: "debug"}))
	username := "test"
	_, err := s.User().GetByUsername(username)
	assert.EqualError(t, err, domain.ErrRecordNotFound.Error())
	u := &models.User{
		Username:          "testuser",
		EncryptedPassword: "encrypted_password",
		Email:             "email@example.com",
		Role:              "user",
	}
	u.Username = username
	// Create a new user
	err = s.User().Create(u)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}
