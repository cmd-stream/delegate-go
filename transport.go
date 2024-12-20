package delegate

import (
	"context"
	"net"
	"time"

	"github.com/cmd-stream/base-go"
)

// Transport is a common transport for the client and server Delegates.
//
// The sent data can be buffered, so there is a Flush() method.
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

// ClienTransport is a transport for the client Delegate.
//
// It is used by the delegate to send commands and receive results.
type ClienTransport[T any] interface {
	Transport[base.Cmd[T], base.Result]
	ReceiveServerInfo() (info ServerInfo, err error)
	ReceiveServerSettings() (settings ServerSettings, err error)
	ApplyServerSettings(settings ServerSettings)
}

// ServerTransportFactory is a factory which creates a Transport for the
// server Delegate.
type ServerTransportFactory[T any] interface {
	New(conn net.Conn) ServerTransport[T]
}

// ServerTransport is a transport for the server Delegate.
//
// It is used by the Delegate to receive commands and send results.
type ServerTransport[T any] interface {
	Transport[base.Result, base.Cmd[T]]
	SendServerInfo(info ServerInfo) error
	SendServerSettings(settings ServerSettings) error
}

// ServerTransportHandler handles the ServerTransport.
//
// It receives, executes commands and sends back results.
type ServerTransportHandler[T any] interface {
	Handle(ctx context.Context, transport ServerTransport[T]) error
}
