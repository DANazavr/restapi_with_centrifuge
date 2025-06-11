package services

import (
	"context"
	"database/sql"

	"github.com/DANazavr/RATest/internal/domain"
	"github.com/DANazavr/RATest/internal/domain/models"
	"github.com/DANazavr/RATest/internal/log"
	"github.com/DANazavr/RATest/internal/store"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	ctx    context.Context
	logger *log.Log
	store  store.Store
}

func NewUserService(ctx context.Context, store store.Store, logger *log.Log) *UserService {
	return &UserService{
		ctx:    ctx,
		store:  store,
		logger: logger.WithComponent("services/user"),
	}
}

func (us *UserService) UsersCreate(ctx context.Context, user *models.User) error {
	if err := us.Validate(user); err != nil {
		return err
	}
	if err := us.BeforeCreate(user); err != nil {
		return err
	}
	if err := us.store.User().Create(user); err != nil {
		return err
	}
	us.Sanitize(user)
	return nil
}

func (us *UserService) UsersGetByUsername(username string) (*models.User, error) {
	user, err := us.store.User().GetByUsername(username)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRecordNotFound
	} else if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrRecordNotFound
	}
	return user, nil
}

func (us *UserService) UsersGetById(id int) (*models.User, error) {
	user, err := us.store.User().GetById(id)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRecordNotFound
	} else if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrRecordNotFound
	}
	return user, nil
}

func (us *UserService) UsersGet() ([]*models.User, error) {
	user, err := us.store.User().Get()
	if err == sql.ErrNoRows {
		return nil, domain.ErrRecordNotFound
	} else if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrRecordNotFound
	}
	return user, nil
}

func (us *UserService) Validate(u *models.User) error {
	return validation.ValidateStruct(u, validation.Field(&u.Username, validation.Required, validation.Length(3, 20)),
		validation.Field(&u.Password, validation.By(func(cond bool) validation.RuleFunc {
			return func(value interface{}) error {
				if cond {
					return validation.Validate(value, validation.Required)
				}
				return nil
			}
		}((u.EncryptedPassword == ""))), validation.Length(6, 20)),
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Role, validation.Required, validation.In("user", "admin")),
	)
}

func (us *UserService) BeforeCreate(u *models.User) error {
	if len(u.Password) > 0 {
		enc, err := us.encryptString(u.Password)
		if err != nil {
			return nil
		}
		u.EncryptedPassword = enc
	}
	return nil
}

func (us *UserService) Sanitize(u *models.User) {
	u.Password = ""
}

func (us *UserService) ComparePassword(u *models.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password)) == nil
}

func (us *UserService) encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
