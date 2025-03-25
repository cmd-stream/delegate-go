package dcln

import (
	"sync"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

const (
	KeepaliveTime  = 3 * time.Second
	KeepaliveIntvl = time.Second
)

// NewKeepalive creates a new KeepaliveDelegate.
func NewKeepalive[T any](d base.ClientDelegate[T], ops ...SetKeepaliveOption) (
	kd KeepaliveDelegate[T]) {
	kd.options = KeepaliveOptions{
		KeepaliveTime:  KeepaliveTime,
		KeepaliveIntvl: KeepaliveIntvl,
	}
	ApplyKeepAlive(ops, &kd.options)
	kd.ClientDelegate = d
	kd.alive = make(chan struct{})
	kd.done = make(chan struct{})
	return
}

// KeepaliveDelegate implements the base.ClientDelegate interface.
//
// When there are no Commands to send, it initiates a Ping-Pong exchange with
// the server. It sends a Ping Command and expects a Pong Result, both
// represented as a single zero byte (like a ball being passed).
type KeepaliveDelegate[T any] struct {
	base.ClientDelegate[T]
	alive   chan struct{}
	done    chan struct{}
	options KeepaliveOptions
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

func (d KeepaliveDelegate[T]) Keepalive(muSn *sync.Mutex) {
	go keepalive[T](d, muSn)
}

func keepalive[T any](d KeepaliveDelegate[T], muSn *sync.Mutex) {
	timer := time.NewTimer(d.options.KeepaliveTime)
	for {
		select {
		case <-d.done:
			return
		case <-timer.C:
			ping(muSn, 0, d) // nothing to do if ping returns an error, TODO
			timer.Reset(d.options.KeepaliveIntvl)
		case <-d.alive:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(d.options.KeepaliveTime)
		}
	}
}

func ping[T any](muSn *sync.Mutex, seq base.Seq, d KeepaliveDelegate[T]) (err error) {
	muSn.Lock()
	if err = d.SetSendDeadline(time.Time{}); err != nil {
		muSn.Unlock()
		return
	}
	if err = d.Send(seq, delegate.PingCmd[T]{}); err != nil {
		muSn.Unlock()
		return
	}
	muSn.Unlock()
	return d.ClientDelegate.Flush()
}
