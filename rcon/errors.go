package rcon

import "errors"

var ErrAuthFail = errors.New("failed to authenticate, wrong password")
var ErrNotRconServer = errors.New("connected server is not an RCON server")
