package dsmock

import (
	"context"

	dser "github.com/cmd-stream/delegate-go/server"
	"github.com/ymz-ncnk/mok"
)

type HandleFn func(ctx context.Context, transport dser.Transport[any]) error

func NewTransportHandler() TransportHandler {
	return TransportHandler{
		Mock: mok.New("TransportHandler"),
	}
}

type TransportHandler struct {
	*mok.Mock
}

func (mock TransportHandler) RegisterHandle(fn HandleFn) TransportHandler {
	mock.Register("Handle", fn)
	return mock
}

func (mock TransportHandler) Handle(ctx context.Context,
	transport dser.Transport[any]) (err error) {
	vals, err := mock.Call("Handle", mok.SafeVal[context.Context](ctx),
		mok.SafeVal[dser.Transport[any]](transport))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
