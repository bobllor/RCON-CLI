package internal

import "errors"

var ErrStartNotAllowed = errors.New("Operation not allowed, use: rcon serve start")
