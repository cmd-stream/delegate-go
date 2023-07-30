package mock

import (
	"context"
	"reflect"

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
	var ctxVal reflect.Value
	if ctx == nil {
		ctxVal = reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem())
	} else {
		ctxVal = reflect.ValueOf(ctx)
	}
	var transportVal reflect.Value
	if transport == nil {
		transportVal = reflect.Zero(reflect.TypeOf((*delegate.ServerTransport[any])(nil)).Elem())
	} else {
		transportVal = reflect.ValueOf(transport)
	}
	vals, err := mock.Call("Handle", ctxVal, transportVal)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
