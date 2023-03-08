package errors

import "errors"

var (
	ErrNoRowsInResultSet = errors.New("sql: no rows in result set")
)
