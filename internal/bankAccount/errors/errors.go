package errors

import "github.com/pkg/errors"

var (
	ErrUnknownEventType     = errors.New("unknown event type")
	ErrInvalidBalanceAmount = errors.New("invalid amount")
)
