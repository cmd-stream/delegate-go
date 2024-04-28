package mock

import (
	"context"

	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

func NewTransportHandler() TransportHandler {
	return TransportHandler{
		Mock: mok.New("TransportHandler"),
	}
}

type TransportHandler struct {
	*mok.Mock
}

func (mock TransportHandler) RegisterHandle(
	fn func(ctx context.Context, transport delegate.ServerTransport[any]) error) TransportHandler {
	mock.Register("Handle", fn)
	return mock
}

func (mock TransportHandler) Handle(ctx context.Context,
	transport delegate.ServerTransport[any]) (err error) {
	vals, err := mock.Call("Handle", mok.SafeVal[context.Context](ctx),
		mok.SafeVal[delegate.ServerTransport[any]](transport))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
