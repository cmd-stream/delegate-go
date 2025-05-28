package dcmock

import (
	dcln "github.com/cmd-stream/delegate-go/client"
	"github.com/ymz-ncnk/mok"
)

type NewFn func() (transport dcln.Transport[any], err error)

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

func (mock TransportFactory) New() (transport dcln.Transport[any],
	err error) {
	vals, err := mock.Call("New")
	if err != nil {
		panic(err)
	}
	transport, _ = vals[0].(dcln.Transport[any])
	err, _ = vals[1].(error)
	return
}
