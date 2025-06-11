package domain

import "errors"

var (
	// store errors
	ErrRecordNotFound = errors.New("record not found")

	//server errors
	ErrIncorectUsernameOrPassword = errors.New("incorrect username or password")
	ErrInvalidSigningMethod       = errors.New("invalid signing method")
	ErrEmptyToken                 = errors.New("empty token")
	ErrInvalidToken               = errors.New("invalid token")
	ErrInvalidAuthHeader          = errors.New("invalid auth header")
	ErrInvalidRequestBody         = errors.New("invalid request body")
	ErrInvalidUserID              = errors.New("invalid user ID")
	ErrInternalEnviroment         = errors.New("enviroment key is not set")

	// service errors
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrUserNotFound             = errors.New("user not found")
	ErrNotificationNotFound     = errors.New("notification not found")
	ErrCentrifugePublishFailed  = errors.New("failed to publish notification to Centrifuge")
	ErrCentrifugePresenceFailed = errors.New("failed to get presence")
	// Err
)
