package mock

import (
	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

func NewClienTransportFactory() ClienTransportFactory {
	return ClienTransportFactory{
		Mock: mok.New("ClienTransportFactory"),
	}
}

type ClienTransportFactory struct {
	*mok.Mock
}

func (mock ClienTransportFactory) RegisterNew(
	fn func() (transport delegate.ClienTransport[any], err error)) ClienTransportFactory {
	mock.Register("New", fn)
	return mock
}

func (mock ClienTransportFactory) New() (transport delegate.ClienTransport[any],
	err error) {
	vals, err := mock.Call("New")
	if err != nil {
		panic(err)
	}
	transport, _ = vals[0].(delegate.ClienTransport[any])
	err, _ = vals[1].(error)
	return
}
