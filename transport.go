package delegate

import (
	"context"
	"net"
	"time"

	"github.com/cmd-stream/base-go"
)

// Transport is a common transport for the client and server delegates.
type Transport[T, V any] interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	SetSendDeadline(deadline time.Time) error
	Send(seq base.Seq, t T) error
	Flush() error

	SetReceiveDeadline(deadline time.Time) error
	Receive() (seq base.Seq, v V, err error)

	Close() error
}

// ClienTransportFactory is a factory which creates a ClientTransport.
type ClienTransportFactory[T any] interface {
	New() (ClienTransport[T], error)
}

// ClienTransport is a transport for the client delegate.
//
// It is used by the delegate to send Commands and receive Results.
type ClienTransport[T any] interface {
	Transport[base.Cmd[T], base.Result]
	ReceiveServerInfo() (info ServerInfo, err error)
}

// ServerTransportFactory is a factory which creates a Transport for the
// server delegate.
type ServerTransportFactory[T any] interface {
	New(conn net.Conn) ServerTransport[T]
}

// ServerTransport is a transport for the server delegate.
//
// It is used by the delegate to receive Commands and send Results.
type ServerTransport[T any] interface {
	Transport[base.Result, base.Cmd[T]]
	SendServerInfo(info ServerInfo) error
}

// ServerTransportHandler is a handler of the ServerTransport.
type ServerTransportHandler[T any] interface {
	Handle(ctx context.Context, transport ServerTransport[T]) error
}
