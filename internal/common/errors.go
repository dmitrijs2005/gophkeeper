package common

import "errors"

var (

	// repository specific errors
	ErrorNotFound = errors.New("not found")

	// service specific errors
	ErrorInternal     = errors.New("internal error")
	ErrorUnauthorized = errors.New("unauthorized")

	// item-specific errors
	ErrorIncorrectMetadata = errors.New("incorrect metadata")

	ErrInvalidToken = errors.New("invalid token")

	// ErrorAlreadyExists = errors.New("already exists")
	// ErrorValidation    = errors.New("validation error")

	// ErrorInvalidAuthheaderFormat = errors.New("invalid auth header format")

	// ErrorNoUserID              = errors.New("no user id")
	// ErrorLoginAlreadyExists    = errors.New("login already exists")
	// ErrorInvalidLoginFormat    = errors.New("invalid login format")
	// ErrorInvalidPasswordFormat = errors.New("invalid password format")
	// ErrorInvalidLoginPassword  = errors.New("invalid login/password")

	// // in-memory repository specific errors
	// ErrorAlreadyInTranscation = errors.New("already in transaction")
	// ErrorNotInTranscation     = errors.New("not in transaction")
)
