package errors

import "github.com/pkg/errors"

var (
	ErrUnknownEventType         = errors.New("unknown event type")
	ErrInvalidBalanceAmount     = errors.New("invalid amount")
	ErrNotEnoughBalance         = errors.New("balance has not enough balance")
	ErrBankAccountNotFound      = errors.New("bank account not found")
	ErrBankAccountAlreadyExists = errors.New("bank account with given id already exists")
)
