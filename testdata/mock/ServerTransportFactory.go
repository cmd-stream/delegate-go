package mock

import (
	"net"

	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

func NewServerTransportFactory() ServerTransportFactory {
	return ServerTransportFactory{
		Mock: mok.New("ServerTransportFactory"),
	}
}

type ServerTransportFactory struct {
	*mok.Mock
}

func (mock ServerTransportFactory) RegisterNew(
	fn func(conn net.Conn) (transport delegate.ServerTransport[any])) ServerTransportFactory {
	mock.Register("New", fn)
	return mock
}

func (mock ServerTransportFactory) New(conn net.Conn) (transport delegate.ServerTransport[any]) {
	vals, err := mock.Call("New", mok.SafeVal[net.Conn](conn))
	if err != nil {
		panic(err)
	}
	transport, _ = vals[0].(delegate.ServerTransport[any])
	return
}
