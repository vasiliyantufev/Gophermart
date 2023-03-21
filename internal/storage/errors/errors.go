package errors

import "errors"

var (
	ErrNotFunds      = errors.New("There are not enough funds on the account")
	ErrNotRegistered = errors.New("Order is not registered in the billing system")
	ErrNotBalance    = errors.New("No balance for order")
)
