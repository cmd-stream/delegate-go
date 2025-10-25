// Package delegate provides client-server communication abstractions for
// cmd-stream-go.
package delegate

import (
	"net"
	"time"

	"github.com/cmd-stream/core-go"
)

// Transport is a common transport for the client and server delegates.
type Transport[T, V any] interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	SetSendDeadline(deadline time.Time) error
	Send(seq core.Seq, t T) (n int, err error)
	Flush() error

	SetReceiveDeadline(deadline time.Time) error
	Receive() (seq core.Seq, v V, n int, err error)

	Close() error
}
