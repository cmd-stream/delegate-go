package server

import "time"

// Conf configures the Delegate.
//
// SysDataSendDuration determines how long the server will try to send system
// data to the client. If == 0, it wait try forever.
type Conf struct {
	SysDataSendDuration time.Duration
}
