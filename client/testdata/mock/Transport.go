package mock

import (
	"net"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

type (
	LocalAddrFn          func() (addr net.Addr)
	RemoteAddrFn         func() (addr net.Addr)
	ReceiveServerInfoFn  func() (info delegate.ServerInfo, err error)
	SetSendDeadlineFn    func(deadline time.Time) (err error)
	SendFn               func(seq core.Seq, cmd core.Cmd[any]) (n int, err error)
	FlushFn              func() (err error)
	SetReceiveDeadlineFn func(deadline time.Time) (err error)
	ReceiveFn            func() (seq core.Seq, result core.Result, n int, err error)
	CloseFn              func() (err error)
)

func NewTransport() Transport {
	return Transport{
		Mock: mok.New("Transport"),
	}
}

type Transport struct {
	*mok.Mock
}

func (mock Transport) RegisterLocalAddr(fn LocalAddrFn) Transport {
	mock.Register("LocalAddr", fn)
	return mock
}

func (mock Transport) RegisterRemoteAddr(fn RemoteAddrFn) Transport {
	mock.Register("RemoteAddr", fn)
	return mock
}

func (mock Transport) RegisterReceiveServerInfo(fn ReceiveServerInfoFn) Transport {
	mock.Register("ReceiveServerInfo", fn)
	return mock
}

func (mock Transport) RegisterSetSendDeadline(fn SetSendDeadlineFn) Transport {
	mock.Register("SetSendDeadline", fn)
	return mock
}

func (mock Transport) RegisterSend(fn SendFn) Transport {
	mock.Register("Send", fn)
	return mock
}

func (mock Transport) RegisterFlush(fn FlushFn) Transport {
	mock.Register("Flush", fn)
	return mock
}

func (mock Transport) RegisterSetReceiveDeadline(fn SetReceiveDeadlineFn) Transport {
	mock.Register("SetReceiveDeadline", fn)
	return mock
}

func (mock Transport) RegisterReceive(fn ReceiveFn) Transport {
	mock.Register("Receive", fn)
	return mock
}

func (mock Transport) RegisterClose(fn CloseFn) Transport {
	mock.Register("Close", fn)
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
	result, err := mock.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = result[0].(net.Addr)
	return
}

func (mock Transport) ReceiveServerInfo() (info delegate.ServerInfo,
	err error,
) {
	vals, err := mock.Call("ReceiveServerInfo")
	if err != nil {
		panic(err)
	}
	info = vals[0].(delegate.ServerInfo)
	err, _ = vals[1].(error)
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

func (mock Transport) Send(seq core.Seq, cmd core.Cmd[any]) (n int,
	err error,
) {
	result, err := mock.Call("Send", seq, mok.SafeVal[core.Cmd[any]](cmd))
	if err != nil {
		panic(err)
	}
	n = result[0].(int)
	err, _ = result[1].(error)
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

func (mock Transport) Receive() (seq core.Seq, result core.Result, n int,
	err error,
) {
	vals, err := mock.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq = vals[0].(core.Seq)
	result, _ = vals[1].(core.Result)
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
