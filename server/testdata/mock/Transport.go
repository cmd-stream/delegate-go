package mock

import (
	"net"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

type LocalAddrFn func() (addr net.Addr)
type RemoteAddrFn func() (addr net.Addr)
type SetSendDeadlineFn func(deadline time.Time) (err error)
type SendFn func(seq core.Seq, result core.Result) (n int, err error)
type FlushFn func() (err error)
type SetReceiveDeadlineFn func(deadline time.Time) (err error)
type ReceiveFn func() (seq core.Seq, cmd core.Cmd[any], n int, err error)
type CloseFn func() (err error)
type SendServerInfo func(info delegate.ServerInfo) (err error)

func NewTransport() Transport {
	return Transport{
		Mock: mok.New("Transport"),
	}
}

type Transport struct {
	*mok.Mock
}

func (mock Transport) RegisterClose(fn CloseFn) Transport {
	mock.Register("Close", fn)
	return mock
}

func (mock Transport) RegisterLocalAddr(fn LocalAddrFn) Transport {
	mock.Register("LocalAddr", fn)
	return mock
}

func (mock Transport) RegisterRemoteAddr(fn RemoteAddrFn) Transport {
	mock.Register("RemoteAddr", fn)
	return mock
}

func (mock Transport) RegisterSetSendDeadline(fn SetSendDeadlineFn) Transport {
	mock.Register("SetSendDeadline", fn)
	return mock
}

func (mock Transport) RegisterNSetSendDeadline(n int,
	fn SetSendDeadlineFn) Transport {
	mock.RegisterN("SetSendDeadline", n, fn)
	return mock
}

func (mock Transport) RegisterSend(fn SendFn) Transport {
	mock.Register("Send", fn)
	return mock
}

func (mock Transport) RegisterNSend(n int, fn SendFn) Transport {
	mock.RegisterN("Send", n, fn)
	return mock
}

func (mock Transport) RegisterFlush(fn FlushFn) Transport {
	mock.Register("Flush", fn)
	return mock
}

func (mock Transport) RegisterNFlush(n int, fn FlushFn) Transport {
	mock.RegisterN("Flush", n, fn)
	return mock
}

func (mock Transport) RegisterSetReceiveDeadline(fn SetReceiveDeadlineFn) Transport {
	mock.Register("SetReceiveDeadline", fn)
	return mock
}

func (mock Transport) RegisterNSetReceiveDeadline(n int,
	fn SetReceiveDeadlineFn) Transport {
	mock.RegisterN("SetReceiveDeadline", n, fn)
	return mock
}

func (mock Transport) RegisterReceive(fn ReceiveFn) Transport {
	mock.Register("Receive", fn)
	return mock
}

func (mock Transport) RegisterSendServerInfo(fn SendServerInfo) Transport {
	mock.Register("SendServerInfo", fn)
	return mock
}

func (mock Transport) LocalAddr() (addr net.Addr) {
	vals, err := mock.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (mock Transport) RemoteAddr() (addr net.Addr) {
	vals, err := mock.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (mock Transport) SendServerInfo(info delegate.ServerInfo) (err error) {
	vals, err := mock.Call("SendServerInfo", info)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (mock Transport) SetSendDeadline(deadline time.Time) (err error) {
	result, err := mock.Call("SetSendDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock Transport) Send(seq core.Seq, result core.Result) (n int,
	err error) {
	vals, err := mock.Call("Send", seq, mok.SafeVal[core.Result](result))
	if err != nil {
		panic(err)
	}
	n = vals[0].(int)
	err, _ = vals[1].(error)
	return
}

func (mock Transport) Flush() (err error) {
	result, err := mock.Call("Flush")
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock Transport) SetReceiveDeadline(deadline time.Time) (err error) {
	result, err := mock.Call("SetReceiveDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock Transport) Receive() (seq core.Seq, cmd core.Cmd[any], n int,
	err error) {
	vals, err := mock.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq = vals[0].(core.Seq)
	cmd, _ = vals[1].(core.Cmd[any])
	n = vals[2].(int)
	err, _ = vals[3].(error)
	return
}

func (mock Transport) Close() (err error) {
	vals, err := mock.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
