package dmock

import (
	"net"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

func NewClienTransport() ClienTransport {
	return ClienTransport{
		Mock: mok.New("ClienTransport"),
	}
}

type ClienTransport struct {
	*mok.Mock
}

func (mock ClienTransport) RegisterLocalAddr(
	fn func() (addr net.Addr)) ClienTransport {
	mock.Register("LocalAddr", fn)
	return mock
}

func (mock ClienTransport) RegisterRemoteAddr(
	fn func() (addr net.Addr)) ClienTransport {
	mock.Register("RemoteAddr", fn)
	return mock
}

func (mock ClienTransport) RegisterReceiveServerInfo(
	fn func() (info delegate.ServerInfo, err error),
) ClienTransport {
	mock.Register("ReceiveServerInfo", fn)
	return mock
}

func (mock ClienTransport) RegisterSetSendDeadline(
	fn func(deadline time.Time) (err error)) ClienTransport {
	mock.Register("SetSendDeadline", fn)
	return mock
}

func (mock ClienTransport) RegisterSend(
	fn func(seq base.Seq, cmd base.Cmd[any]) (err error)) ClienTransport {
	mock.Register("Send", fn)
	return mock
}

func (mock ClienTransport) RegisterFlush(fn func() (err error)) ClienTransport {
	mock.Register("Flush", fn)
	return mock
}

func (mock ClienTransport) RegisterSetReceiveDeadline(
	fn func(deadline time.Time) (err error)) ClienTransport {
	mock.Register("SetReceiveDeadline", fn)
	return mock
}

func (mock ClienTransport) RegisterReceive(
	fn func() (seq base.Seq, result base.Result, err error)) ClienTransport {
	mock.Register("Receive", fn)
	return mock
}

func (mock ClienTransport) RegisterClose(
	fn func() (err error)) ClienTransport {
	mock.Register("Close", fn)
	return mock
}

func (mock ClienTransport) LocalAddr() (addr net.Addr) {
	vals, err := mock.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (mock ClienTransport) RemoteAddr() (addr net.Addr) {
	result, err := mock.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = result[0].(net.Addr)
	return
}

func (mock ClienTransport) ReceiveServerInfo() (info delegate.ServerInfo, err error) {
	vals, err := mock.Call("ReceiveServerInfo")
	if err != nil {
		panic(err)
	}
	info = vals[0].(delegate.ServerInfo)
	err, _ = vals[1].(error)
	return
}

func (mock ClienTransport) SetSendDeadline(deadline time.Time) (err error) {
	result, err := mock.Call("SetSendDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ClienTransport) Send(seq base.Seq, cmd base.Cmd[any]) (err error) {
	result, err := mock.Call("Send", seq, mok.SafeVal[base.Cmd[any]](cmd))
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ClienTransport) Flush() (err error) {
	result, err := mock.Call("Flush")
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ClienTransport) SetReceiveDeadline(deadline time.Time) (err error) {
	result, err := mock.Call("SetReceiveDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ClienTransport) Receive() (seq base.Seq, result base.Result, err error) {
	vals, err := mock.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq, _ = vals[0].(base.Seq)
	result, _ = vals[1].(base.Result)
	err, _ = vals[2].(error)
	return
}

func (mock ClienTransport) Close() (err error) {
	vals, err := mock.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
