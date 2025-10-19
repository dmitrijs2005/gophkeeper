package client

import "errors"

// ErrUnavailable indicates that the server (or network) is unreachable.
// Callers may choose to fall back to offline flows when this error occurs.
var ErrUnavailable = errors.New("server unavailable")

// ErrUnauthorized indicates that authentication failed or credentials expired.
// Clients should prompt for re-authentication or refresh tokens.
var ErrUnauthorized = errors.New("unauthorized")

// ErrLocalDataNotAvailable indicates that required local state (e.g., cached
// entries or keys for offline mode) is missing or unreadable.
var ErrLocalDataNotAvailable = errors.New("local data unavailable")
