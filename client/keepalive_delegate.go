package client

import (
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// NewKeepalive creates a new KeepaliveDelegate.
func NewKeepalive[T any](conf Conf,
	d base.ClientDelegate[T]) KeepaliveDelegate[T] {
	dlgt := KeepaliveDelegate[T]{
		ClientDelegate: d,
		conf:           conf,
		alive:          make(chan struct{}),
		done:           make(chan struct{}),
	}
	go keepalive[T](dlgt)
	return dlgt
}

// KeepaliveDelegate is an implementation of the base.ClientDelegate interface.
//
// KeepaliveDelegate is a delegate which keeps the connection alive. When there
// is no commands to send, it starts Ping-Pong with the server - sends the Ping
// command and receives the Pong result, both of which are transfered as a 0
// (like a ball) byte.
type KeepaliveDelegate[T any] struct {
	base.ClientDelegate[T]
	conf  Conf
	alive chan struct{}
	done  chan struct{}
}

func (d KeepaliveDelegate[T]) Receive() (seq base.Seq, result base.Result,
	err error) {
Start:
	seq, result, err = d.ClientDelegate.Receive()
	if err != nil {
		return
	}
	if _, ok := result.(delegate.PongResult); ok {
		goto Start
	}
	return
}

func (d KeepaliveDelegate[T]) Flush() (err error) {
	if err = d.ClientDelegate.Flush(); err != nil {
		return
	}
	select {
	case d.alive <- struct{}{}:
	default:
	}
	return
}

func (d KeepaliveDelegate[T]) Close() (err error) {
	if err = d.ClientDelegate.Close(); err != nil {
		return
	}
	close(d.done)
	return
}

func keepalive[T any](d KeepaliveDelegate[T]) {
	timer := time.NewTimer(d.conf.KeepaliveTime)
	for {
		select {
		case <-d.done:
			return
		case <-timer.C:
			ping(0, d) // nothing to do if ping returns an error
			timer.Reset(d.conf.KeepaliveIntvl)
		case <-d.alive:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(d.conf.KeepaliveTime)
		}
	}
}

func ping[T any](seq base.Seq, d KeepaliveDelegate[T]) (err error) {
	if err = d.Send(seq, delegate.PingCmd[T]{}); err != nil {
		return
	}
	return d.ClientDelegate.Flush()
}
