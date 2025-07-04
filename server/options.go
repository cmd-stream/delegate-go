package server

import "time"

type Options struct {
	ServerInfoSendDuration time.Duration
}

type SetOption func(o *Options)

// WithServerInfoDuration specifies how long the server will try to send
// ServerInfo to the client. If == 0, it will try forever.
func WithServerInfoSendDuration(d time.Duration) SetOption {
	return func(o *Options) { o.ServerInfoSendDuration = d }
}

func Apply(ops []SetOption, o *Options) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
}
