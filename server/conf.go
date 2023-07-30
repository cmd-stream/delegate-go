package server

import "time"

// Conf is a Delegate configuration.
//
// SysDataSendTimeout determines how long the server will try to send system
// data to the client, if == 0, waits forever.
type Conf struct {
	SysDataSendTimeout time.Duration
}
