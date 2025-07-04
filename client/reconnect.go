package client

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/cmd-stream/core-go"
	cln "github.com/cmd-stream/core-go/client"
	"github.com/cmd-stream/delegate-go"
)

// NewReconnect creates a new ReconnectDelegate.
func NewReconnect[T any](info delegate.ServerInfo, factory TransportFactory[T],
	ops ...SetOption) (d ReconnectDelegate[T], err error) {
	transport, err := factory.New()
	if err != nil {
		return
	}
	Apply(ops, &d.options)
	err = checkServerInfo(d.options.ServerInfoReceiveDuration, transport, info)
	if err != nil {
		return
	}
	var closedFlag uint32
	d.info = info
	d.factory = factory
	d.closedFlag = &closedFlag
	d.transport = &atomic.Value{}
	d.setTransport(transport)
	return
}

// NewReconnectWithoutInfo for tests only.
func NewReconnectWithoutInfo[T any](factory TransportFactory[T],
	closedFlag *uint32,
	transport *atomic.Value,
	options Options,
) ReconnectDelegate[T] {
	return ReconnectDelegate[T]{
		factory:    factory,
		closedFlag: closedFlag,
		transport:  transport,
		options:    options,
	}
}

// ReconnectDelegate implements the core.ClientReconnectDelegate interface.
type ReconnectDelegate[T any] struct {
	info       delegate.ServerInfo
	factory    TransportFactory[T]
	closedFlag *uint32
	transport  *atomic.Value
	options    Options
}

func (d ReconnectDelegate[T]) Options() Options {
	return d.options
}

func (d ReconnectDelegate[T]) LocalAddr() net.Addr {
	return d.Transport().LocalAddr()
}

func (d ReconnectDelegate[T]) RemoteAddr() net.Addr {
	return d.Transport().RemoteAddr()
}

func (d ReconnectDelegate[T]) SetSendDeadline(deadline time.Time) error {
	return d.Transport().SetSendDeadline(deadline)
}

func (d ReconnectDelegate[T]) Send(seq core.Seq, cmd core.Cmd[T]) (n int,
	err error) {
	return d.Transport().Send(seq, cmd)
}

func (d ReconnectDelegate[T]) Flush() error {
	return d.Transport().Flush()
}

func (d ReconnectDelegate[T]) SetReceiveDeadline(deadline time.Time) error {
	return d.Transport().SetReceiveDeadline(deadline)
}

func (d ReconnectDelegate[T]) Receive() (seq core.Seq, result core.Result,
	n int, err error) {
	return d.Transport().Receive()
}

func (d ReconnectDelegate[T]) Close() (err error) {
	err = d.Transport().Close()
	if err != nil {
		return
	}
	if swapped := atomic.CompareAndSwapUint32(d.closedFlag, 0, 1); !swapped {
		panic("can'transport close")
	}
	return
}

func (d ReconnectDelegate[T]) Reconnect() (err error) {
	var transport Transport[T]
Start:
	for {
		if d.closed() {
			return cln.ErrClosed
		}
		transport, err = d.factory.New()
		if err != nil {
			continue
		}
		break
	}
	err = checkServerInfo(d.options.ServerInfoReceiveDuration, transport, d.info)
	if err != nil {
		if err == ErrServerInfoMismatch {
			return
		}
		goto Start
	}
	d.setTransport(transport)
	return
}

func (d ReconnectDelegate[T]) setTransport(transport Transport[T]) {
	d.transport.Store(transport)
}

func (d ReconnectDelegate[T]) Transport() Transport[T] {
	return d.transport.Load().(Transport[T])
}

func (d ReconnectDelegate[T]) closed() bool {
	return !atomic.CompareAndSwapUint32(d.closedFlag, 0, 0)
}
