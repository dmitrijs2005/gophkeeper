package client

import "errors"

var (
	ErrUnavailable           = errors.New("server unavailable")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrLocalDataNotAvailable = errors.New("local data unavailable")
)
