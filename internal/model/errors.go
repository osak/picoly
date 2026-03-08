package model

import "errors"

var (
	ErrNotFound      = errors.New("ticket not found")
	ErrRaceCondition = errors.New("race condition detected: ticket was updated by another process")
	ErrUnauthorized  = errors.New("unauthorized: insufficient role")
	ErrInvalidStatus = errors.New("invalid status value")
	ErrMissingUserID = errors.New("PICOLY_USER_ID environment variable is not set")
)
