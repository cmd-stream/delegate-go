package dcln

import (
	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/delegate-go"
)

// TransportFactory is a factory which creates a ClientTransport.
type TransportFactory[T any] interface {
	New() (Transport[T], error)
}

// Transport is a transport for the client delegate.
//
// It is used by the delegate to send Commands and receive Results.
type Transport[T any] interface {
	delegate.Transport[base.Cmd[T], base.Result]
	ReceiveServerInfo() (info delegate.ServerInfo, err error)
}
