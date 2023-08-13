package server

import "errors"

// ErrEmptyInfo happens on Delegate creation, when ServerInfo is empty.
var ErrEmptyInfo = errors.New("empty info")
