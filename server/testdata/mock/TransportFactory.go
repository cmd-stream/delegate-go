package mock

import (
	"net"

	dsrv "github.com/cmd-stream/delegate-go/server"
	"github.com/ymz-ncnk/mok"
)

type NewFn func(conn net.Conn) (transport dsrv.Transport[any])

func NewTransportFactory() TransportFactory {
	return TransportFactory{
		Mock: mok.New("TransportFactory"),
	}
}

type TransportFactory struct {
	*mok.Mock
}

func (mock TransportFactory) RegisterNew(fn NewFn) TransportFactory {
	mock.Register("New", fn)
	return mock
}

func (mock TransportFactory) New(conn net.Conn) (
	transport dsrv.Transport[any],
) {
	vals, err := mock.Call("New", mok.SafeVal[net.Conn](conn))
	if err != nil {
		panic(err)
	}
	transport, _ = vals[0].(dsrv.Transport[any])
	return
}
