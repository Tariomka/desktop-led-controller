package common

import "errors"

var (
	ErrOutsideBasePath = errors.New("navigating outside base path not allowed")
)
