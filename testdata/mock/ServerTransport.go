package mock

import (
	"net"
	"reflect"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

func NewServerTransport() ServerTransport {
	return ServerTransport{
		Mock: mok.New("ServerTransport"),
	}
}

type ServerTransport struct {
	*mok.Mock
}

func (mock ServerTransport) RegisterClose(
	fn func() (err error)) ServerTransport {
	mock.Register("Close", fn)
	return mock
}

func (mock ServerTransport) RegisterLocalAddr(
	fn func() (addr net.Addr)) ServerTransport {
	mock.Register("LocalAddr", fn)
	return mock
}

func (mock ServerTransport) RegisterRemoteAddr(
	fn func() (addr net.Addr)) ServerTransport {
	mock.Register("RemoteAddr", fn)
	return mock
}

func (mock ServerTransport) RegisterSetSendDeadline(
	fn func(deadline time.Time) (err error)) ServerTransport {
	mock.Register("SetSendDeadline", fn)
	return mock
}

func (mock ServerTransport) RegisterNSetSendDeadline(n int,
	fn func(deadline time.Time) (err error)) ServerTransport {
	mock.RegisterN("SetSendDeadline", n, fn)
	return mock
}

func (mock ServerTransport) RegisterSend(
	fn func(seq base.Seq, result base.Result) (err error)) ServerTransport {
	mock.Register("Send", fn)
	return mock
}

func (mock ServerTransport) RegisterNSend(n int,
	fn func(seq base.Seq, result base.Result) (err error)) ServerTransport {
	mock.RegisterN("Send", n, fn)
	return mock
}

func (mock ServerTransport) RegisterFlush(fn func() (err error)) ServerTransport {
	mock.Register("Flush", fn)
	return mock
}

func (mock ServerTransport) RegisterNFlush(n int,
	fn func() (err error)) ServerTransport {
	mock.RegisterN("Flush", n, fn)
	return mock
}

func (mock ServerTransport) RegisterSetReceiveDeadline(
	fn func(deadline time.Time) (err error)) ServerTransport {
	mock.Register("SetReceiveDeadline", fn)
	return mock
}

func (mock ServerTransport) RegisterNSetReceiveDeadline(n int,
	fn func(deadline time.Time) (err error)) ServerTransport {
	mock.RegisterN("SetReceiveDeadline", n, fn)
	return mock
}

func (mock ServerTransport) RegisterReceive(
	fn func() (seq base.Seq, cmd base.Cmd[any], err error)) ServerTransport {
	mock.Register("Receive", fn)
	return mock
}

func (mock ServerTransport) RegisterSendServerInfo(
	fn func(info delegate.ServerInfo) (err error)) ServerTransport {
	mock.Register("SendServerInfo", fn)
	return mock
}

func (mock ServerTransport) RegisterSendServerSettings(
	fn func(settings delegate.ServerSettings) (err error)) ServerTransport {
	mock.Register("SendServerSettings", fn)
	return mock
}

func (mock ServerTransport) LocalAddr() (addr net.Addr) {
	vals, err := mock.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (mock ServerTransport) RemoteAddr() (addr net.Addr) {
	vals, err := mock.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (mock ServerTransport) SendServerInfo(info delegate.ServerInfo) (err error) {
	vals, err := mock.Call("SendServerInfo", info)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (mock ServerTransport) SendServerSettings(settings delegate.ServerSettings) (
	err error) {
	vals, err := mock.Call("SendServerSettings", settings)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (mock ServerTransport) SetSendDeadline(deadline time.Time) (err error) {
	result, err := mock.Call("SetSendDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ServerTransport) Send(seq base.Seq, result base.Result) (err error) {
	var resultVal reflect.Value
	if result == nil {
		resultVal = reflect.Zero(reflect.TypeOf((*base.Result)(nil)).Elem())
	} else {
		resultVal = reflect.ValueOf(result)
	}
	vals, err := mock.Call("Send", seq, resultVal)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (mock ServerTransport) Flush() (err error) {
	result, err := mock.Call("Flush")
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ServerTransport) SetReceiveDeadline(deadline time.Time) (err error) {
	result, err := mock.Call("SetReceiveDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (mock ServerTransport) Receive() (seq base.Seq, cmd base.Cmd[any], err error) {
	vals, err := mock.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq, _ = vals[0].(base.Seq)
	cmd, _ = vals[1].(base.Cmd[any])
	err, _ = vals[2].(error)
	return
}

func (mock ServerTransport) Close() (err error) {
	vals, err := mock.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
