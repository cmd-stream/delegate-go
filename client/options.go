package dcln

import "time"

type Options struct {
	ServerInfoReceiveDuration time.Duration
}

type SetOption func(o *Options)

// WithServerInfoReceiveDuration sets the duration the client will wait
// for the ServerInfo. If set to 0, the client waits indefinitely.
func WithServerInfoReceiveDuration(d time.Duration) SetOption {
	return func(o *Options) { o.ServerInfoReceiveDuration = d }
}

func Apply(ops []SetOption, o *Options) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
}

type KeepaliveOptions struct {
	KeepaliveTime  time.Duration
	KeepaliveIntvl time.Duration
}

type SetKeepaliveOption func(o *KeepaliveOptions)

// WithKeepaliveTime sets the inactivity period after which the client
// starts sending Ping Commands to the server if no Commands have been sent.
func WithKeepaliveTime(d time.Duration) SetKeepaliveOption {
	return func(o *KeepaliveOptions) { o.KeepaliveTime = d }
}

// WithKeepaliveIntvl sets the interval between consecutive Ping Commands
// sent by the client.
func WithKeepaliveIntvl(d time.Duration) SetKeepaliveOption {
	return func(o *KeepaliveOptions) { o.KeepaliveIntvl = d }
}

func ApplyKeepAlive(ops []SetKeepaliveOption, o *KeepaliveOptions) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
}
