package client

import "errors"

// ErrServerInfoMismatch happens when the ServerInfo of the client and server
// does not match.
var ErrServerInfoMismatch = errors.New("server info mismatch")
