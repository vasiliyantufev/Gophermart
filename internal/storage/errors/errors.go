package errors

import "errors"

var (
	ErrNoRowsInResultSet = errors.New("sql: no rows in result set")
	ErrNotFunds          = errors.New("There are not enough funds on the account")
)
