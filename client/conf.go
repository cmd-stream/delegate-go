package client

import "time"

// Conf configures the Delegate.
//
//   - SysDataReceiveDuration specifies how long the client will wait for the
//     server's system data. If set to 0, the client will wait indefinitely.
//   - KeepaliveTime defines the period of inactivity after which the client
//     will start sending Ping Commands to the server if no Commands have been
//     sent.
//   - KeepaliveIntvl sets the time interval between consecutive Ping Commands
//     sent by the client.
type Conf struct {
	SysDataReceiveDuration time.Duration
	KeepaliveTime          time.Duration
	KeepaliveIntvl         time.Duration
}
