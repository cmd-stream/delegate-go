package dsrv

import (
	"context"
	"net"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// TransportFactory is a factory which creates a Transport for the
// server delegate.
type TransportFactory[T any] interface {
	New(conn net.Conn) Transport[T]
}

// Transport is a transport for the server delegate.
//
// It is used by the delegate to receive Commands and send Results.
type Transport[T any] interface {
	delegate.Transport[base.Result, base.Cmd[T]]
	SendServerInfo(info delegate.ServerInfo) error
}

// TransportHandler is a handler of the Transport.
type TransportHandler[T any] interface {
	Handle(ctx context.Context, transport Transport[T]) error
}
