package shared

import "errors"

var (

	// common errors
	ErrorNotFound = errors.New("not found")

	// auth-specific errors
	ErrorInvalidToken = errors.New("invalid token")

	ErrorAlreadyExists = errors.New("already exists")
	ErrorValidation    = errors.New("validation error")

	ErrorInvalidAuthheaderFormat = errors.New("invalid auth header format")

	ErrorNoUserID              = errors.New("no user id")
	ErrorLoginAlreadyExists    = errors.New("login already exists")
	ErrorInvalidLoginFormat    = errors.New("invalid login format")
	ErrorInvalidPasswordFormat = errors.New("invalid password format")
	ErrorInvalidLoginPassword  = errors.New("invalid login/password")

	// order-specific errors
	ErrorNoOrderNumberSpecified   = errors.New("no order number specified")
	ErrorInvalidOrderNumberFormat = errors.New("invalid order number format")
	ErrorOrderDoesNotExist        = errors.New("order does not exist")
	ErrorOrderAlreadyExists       = errors.New("order already exists")

	// balance-specific errors
	ErrorInsufficientBalance = errors.New("insufficient balance")

	// in-memory repository specific errors
	ErrorAlreadyInTranscation = errors.New("already in transaction")
	ErrorNotInTranscation     = errors.New("not in transaction")
)
