package mock

import (
	"net"
	"reflect"

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
	var connVal reflect.Value
	if conn == nil {
		connVal = reflect.Zero(reflect.TypeOf((*net.Conn)(nil)).Elem())
	} else {
		connVal = reflect.ValueOf(conn)
	}
	vals, err := mock.Call("New", connVal)
	if err != nil {
		panic(err)
	}
	transport, _ = vals[0].(delegate.ServerTransport[any])
	return
}
