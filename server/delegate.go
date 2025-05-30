package dsrv

import (
	"context"
	"net"
	"time"

	"github.com/cmd-stream/delegate-go"
)

// New creates a new Delegate.
//
// Panics with ErrEmptyInfo if ServerInfo is empty.
func New[T any](info delegate.ServerInfo, factory TransportFactory[T],
	handler TransportHandler[T],
	ops ...SetOption,
) (d Delegate[T]) {
	if len(info) == 0 {
		panic(ErrEmptyInfo)
	}
	Apply(ops, &d.options)
	d.info = info
	d.factory = factory
	d.handler = handler
	return
}

// Delegate implements the base.ServerDelegate interface.
//
// It initializes the connection by sending ServerInfo to the client.
type Delegate[T any] struct {
	info    delegate.ServerInfo
	factory TransportFactory[T]
	handler TransportHandler[T]
	options Options
}

func (d Delegate[T]) Handle(ctx context.Context, conn net.Conn) (err error) {
	transport := d.factory.New(conn)
	err = d.sendServerInfo(transport)
	if err != nil {
		if err := transport.Close(); err != nil {
			panic(err)
		}
		return err
	}
	return d.handler.Handle(ctx, transport)
}

func (d Delegate[T]) sendServerInfo(transport Transport[T]) (err error) {
	if d.options.ServerInfoSendDuration != 0 {
		deadline := time.Now().Add(d.options.ServerInfoSendDuration)
		if err = transport.SetSendDeadline(deadline); err != nil {
			return
		}
	}
	return transport.SendServerInfo(d.info)
}
