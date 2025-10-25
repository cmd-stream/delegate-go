// Package client provides client-side implementations for the delegate
// abstraction of the cmd-stream-go library.
//
// It defines several Delegate types that implement the core.ClientDelegate
// and core.ClientReconnectDelegate interfaces.
//
// Key delegates:
//
//   - Delegate: basic client delegate that receives ServerInfo.
//   - KeepaliveDelegate: extends Delegate with a ping-pong mechanism to keep
//     the connection alive when no Commands are pending.
//   - ReconnectDelegate: extends Delegate with automatic reconnect logic
//     when the connection to the server is lost.
//
// All delegates rely on a pluggable Transport for data exchange and support
// configurable options such as send/receive deadlines.
package client

import (
	"bytes"
	"net"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/delegate-go"
)

// New creates a new Delegate.
//
// The Delegate expects to receive ServerInfo from the server upon creation.
//
// Returns ErrServerInfoMismatch if the received ServerInfo does not match
// the specified one.
func New[T any](info delegate.ServerInfo, transport Transport[T],
	ops ...SetOption,
) (d Delegate[T], err error) {
	Apply(ops, &d.options)
	err = checkServerInfo(d.options.ServerInfoReceiveDuration, transport, info)
	if err != nil {
		return
	}
	d.transport = transport
	return
}

// NewWithoutInfo for tests only.
func NewWithoutInfo[T any](transport Transport[T]) (d Delegate[T]) {
	d.transport = transport
	return
}

// Delegate implements the core.ClientDelegate interface.
type Delegate[T any] struct {
	transport Transport[T]
	options   Options
}

func (d Delegate[T]) Options() Options {
	return d.options
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

func (d Delegate[T]) Send(seq core.Seq, cmd core.Cmd[T]) (n int, err error) {
	return d.transport.Send(seq, cmd)
}

func (d Delegate[T]) Flush() error {
	return d.transport.Flush()
}

func (d Delegate[T]) SetReceiveDeadline(deadline time.Time) error {
	return d.transport.SetReceiveDeadline(deadline)
}

func (d Delegate[T]) Receive() (seq core.Seq, result core.Result, n int,
	err error,
) {
	return d.transport.Receive()
}

func (d Delegate[T]) Close() error {
	return d.transport.Close()
}

func checkServerInfo[T any](timeout time.Duration,
	transport Transport[T],
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
	return transport.SetReceiveDeadline(time.Time{})
}

func calcDeadline(duration time.Duration) (deadline time.Time) {
	if duration != 0 {
		deadline = time.Now().Add(duration)
	}
	return
}
