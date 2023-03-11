package api

import "errors"

var (
	ErrNoRowsInResultSet = errors.New("sql: no rows in result set")
)
