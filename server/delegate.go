package server

import (
	"context"
	"net"
	"time"

	"github.com/cmd-stream/delegate-go"
)

// New creates a new Delegate initialized by the transport factory and handler.
//
// If ServerInfo is empty, it panics with ErrEmptyInfo.
func New[T any](conf Conf, info delegate.ServerInfo,
	settings delegate.ServerSettings,
	factory delegate.ServerTransportFactory[T],
	handler delegate.ServerTransportHandler[T],
) Delegate[T] {
	if len(info) == 0 {
		panic(ErrEmptyInfo)
	}
	return Delegate[T]{conf, info, settings, factory, handler}
}

// Delegate is an implementation of the base.ServerDelegate interface.
//
// It initialize the connection by sending system data (ServerInfo and
// ServerSettins) to the client.
type Delegate[T any] struct {
	conf     Conf
	info     delegate.ServerInfo
	settings delegate.ServerSettings
	factory  delegate.ServerTransportFactory[T]
	handler  delegate.ServerTransportHandler[T]
}

func (h Delegate[T]) Handle(ctx context.Context, conn net.Conn) (
	err error) {
	transport := h.factory.New(conn)
	err = h.sendServerInfo(transport)
	if err != nil {
		if err := transport.Close(); err != nil {
			panic(err)
		}
		return err
	}
	err = h.sendServerSettings(transport)
	if err != nil {
		if err := transport.Close(); err != nil {
			panic(err)
		}
		return err
	}
	return h.handler.Handle(ctx, transport)
}

func (h Delegate[T]) sendServerInfo(transport delegate.ServerTransport[T]) (
	err error) {
	if h.conf.SysDataSendDuration != 0 {
		deadline := time.Now().Add(h.conf.SysDataSendDuration)
		if err = transport.SetSendDeadline(deadline); err != nil {
			return
		}
	}
	return transport.SendServerInfo(h.info)
}

func (h Delegate[T]) sendServerSettings(transport delegate.ServerTransport[T]) (
	err error) {
	if h.conf.SysDataSendDuration != 0 {
		deadline := time.Now().Add(h.conf.SysDataSendDuration)
		if err = transport.SetSendDeadline(deadline); err != nil {
			return
		}
	}
	return transport.SendServerSettings(h.settings)
}
