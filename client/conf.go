package client

import "time"

// Conf is a Delegate configuration.
//
// SysDataReceiveTimeout specifies how long the client will wait for the server
// system data, if == 0, waits forever.
//
// KeepaliveTime - if the client has not sent any commands during this time, it
// will start sending Ping commands to the server.
//
// KeepaliveIntvl sets the time interval between Ping commands.
type Conf struct {
	SysDataReceiveTimeout time.Duration
	KeepaliveTime         time.Duration
	KeepaliveIntvl        time.Duration
}
