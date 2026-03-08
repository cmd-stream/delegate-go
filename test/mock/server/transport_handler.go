package server

import (
	"context"

	dsrv "github.com/cmd-stream/delegate-go/server"
	"github.com/ymz-ncnk/mok"
)

type HandleFn func(ctx context.Context, transport dsrv.Transport[any]) error

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
	transport dsrv.Transport[any],
) (err error) {
	vals, err := mock.Call("Handle", ctx, transport)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
