package client

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/cmd-stream/base-go"
	bcln "github.com/cmd-stream/base-go/client"
	"github.com/cmd-stream/delegate-go"
)

// NewReconnect creates a new ReconnectDelegate.
func NewReconnect[T any](conf Conf, info delegate.ServerInfo,
	factory delegate.ClienTransportFactory[T],
) (delegate *ReconnectDelegate[T], err error) {
	transport, err := factory.New()
	if err != nil {
		return
	}
	err = checkServerInfo(conf.SysDataReceiveDuration, transport, info)
	if err != nil {
		return
	}
	err = applyServerSettings(conf.SysDataReceiveDuration, transport)
	if err != nil {
		return
	}
	var closedFlag uint32
	delegate = &ReconnectDelegate[T]{
		conf:       conf,
		info:       info,
		factory:    factory,
		closedFlag: &closedFlag,
	}
	delegate.setTransport(transport)
	return
}

// ReconnectDelegate is an implementation of the base.ClientReconnectDelegate
// interface.
type ReconnectDelegate[T any] struct {
	conf       Conf
	info       delegate.ServerInfo
	transport  atomic.Value
	factory    delegate.ClienTransportFactory[T]
	closedFlag *uint32
}

func (d *ReconnectDelegate[T]) Conf() Conf {
	return d.conf
}

func (d *ReconnectDelegate[T]) LocalAddr() net.Addr {
	return d.getTransport().LocalAddr()
}

func (d *ReconnectDelegate[T]) RemoteAddr() net.Addr {
	return d.getTransport().RemoteAddr()
}

func (d *ReconnectDelegate[T]) SetSendDeadline(deadline time.Time) error {
	return d.getTransport().SetSendDeadline(deadline)
}

func (d *ReconnectDelegate[T]) Send(seq base.Seq, cmd base.Cmd[T]) (err error) {
	return d.getTransport().Send(seq, cmd)
}

func (d *ReconnectDelegate[T]) Flush() error {
	return d.getTransport().Flush()
}

func (d *ReconnectDelegate[T]) SetReceiveDeadline(deadline time.Time) error {
	return d.getTransport().SetReceiveDeadline(deadline)
}

func (d *ReconnectDelegate[T]) Receive() (seq base.Seq, result base.Result,
	err error) {
	return d.getTransport().Receive()
}

func (d *ReconnectDelegate[T]) Close() (err error) {
	err = d.getTransport().Close()
	if err != nil {
		return
	}
	if swapped := atomic.CompareAndSwapUint32(d.closedFlag, 0, 1); !swapped {
		panic("can't close")
	}
	return
}

func (d *ReconnectDelegate[T]) Reconnect() (err error) {
	var transport delegate.ClienTransport[T]
Start:
	for {
		if d.closed() {
			return bcln.ErrClosed
		}
		transport, err = d.factory.New()
		if err != nil {
			continue
		}
		break
	}
	err = checkServerInfo(d.conf.SysDataReceiveDuration, transport, d.info)
	if err != nil {
		if err == ErrServerInfoMismatch {
			return
		}
		goto Start
	}
	err = applyServerSettings(d.conf.SysDataReceiveDuration, transport)
	if err != nil {
		goto Start
	}
	d.setTransport(transport)
	return
}

func (d *ReconnectDelegate[T]) setTransport(
	transport delegate.ClienTransport[T]) {
	d.transport.Store(transport)
}

func (d *ReconnectDelegate[T]) getTransport() delegate.ClienTransport[T] {
	return d.transport.Load().(delegate.ClienTransport[T])
}

func (d *ReconnectDelegate[T]) closed() bool {
	return !atomic.CompareAndSwapUint32(d.closedFlag, 0, 0)
}
