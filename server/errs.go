package server

import "errors"

// ErrEmptyInfo happens on the delegate creation, when the ServerInfo is empty.
var ErrEmptyInfo = errors.New("empty info")
