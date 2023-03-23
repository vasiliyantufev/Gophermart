package errors

import "errors"

var (
	ErrNotFunds      = errors.New("there are not enough funds on the account")
	ErrNotRegistered = errors.New("order is not registered in the billing system")
	ErrNotBalance    = errors.New("no balance for order")
)
