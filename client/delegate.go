package client

import (
	"bytes"
	"net"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// New creates a new Delegate.
//
// When created, the Delegate expects to receive system data (ServerInfo and
// ServerSettings) from the server.
//
// Returns ErrServerInfoMismatch if the received ServerInfo does not match
// the specified one.
func New[T any](conf Conf, info delegate.ServerInfo,
	transport delegate.ClienTransport[T]) (delegate Delegate[T], err error) {
	err = checkServerInfo(conf.SysDataReceiveDuration, transport, info)
	if err != nil {
		return
	}
	err = applyServerSettings(conf.SysDataReceiveDuration, transport)
	if err != nil {
		return
	}
	delegate = Delegate[T]{conf, transport}
	return
}

// Delegate is an implementation of the base.ClientDelegate interface.
type Delegate[T any] struct {
	conf      Conf
	transport delegate.ClienTransport[T]
}

func (d Delegate[T]) Conf() Conf {
	return d.conf
}

func (d Delegate[T]) LocalAddr() net.Addr {
	return d.transport.LocalAddr()
}

func (d Delegate[T]) RemoteAddr() net.Addr {
	return d.transport.RemoteAddr()
}

func (d Delegate[T]) SetSendDeadline(deadline time.Time) error {
	return d.transport.SetSendDeadline(deadline)
}

func (d Delegate[T]) Send(seq base.Seq, cmd base.Cmd[T]) (err error) {
	return d.transport.Send(seq, cmd)
}

func (d Delegate[T]) Flush() error {
	return d.transport.Flush()
}

func (d Delegate[T]) SetReceiveDeadline(deadline time.Time) error {
	return d.transport.SetReceiveDeadline(deadline)
}

func (d Delegate[T]) Receive() (seq base.Seq, result base.Result, err error) {
	return d.transport.Receive()
}

func (d Delegate[T]) Close() error {
	return d.transport.Close()
}

func checkServerInfo[T any](timeout time.Duration,
	transport delegate.ClienTransport[T],
	wantInfo delegate.ServerInfo,
) (err error) {
	err = transport.SetReceiveDeadline(calcDeadline(timeout))
	if err != nil {
		return
	}
	info, err := transport.ReceiveServerInfo()
	if err != nil {
		return
	}
	if !bytes.Equal(info, wantInfo) {
		return ErrServerInfoMismatch
	}
	return
}

func applyServerSettings[T any](timeout time.Duration,
	transport delegate.ClienTransport[T]) (err error) {
	err = transport.SetReceiveDeadline(calcDeadline(timeout))
	if err != nil {
		return
	}
	settings, err := transport.ReceiveServerSettings()
	if err != nil {
		return
	}
	transport.ApplyServerSettings(settings)
	return transport.SetReceiveDeadline(time.Time{})
}

func calcDeadline(duration time.Duration) (deadline time.Time) {
	if duration != 0 {
		deadline = time.Now().Add(duration)
	}
	return
}
