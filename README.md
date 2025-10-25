# delegate-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/delegate-go.svg)](https://pkg.go.dev/github.com/cmd-stream/delegate-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/delegate-go)](https://goreportcard.com/report/github.com/cmd-stream/delegate-go)
[![codecov](https://codecov.io/gh/cmd-stream/delegate-go/graph/badge.svg?token=G8NN40DYJI)](https://codecov.io/gh/cmd-stream/delegate-go)

**delegate-go** facilitates communication between the `cmd-stream-go` client
and server. It provides implementations of the `core.ClientDelegate` and
`core.ServerDelegate` interfaces, both of which rely on an abstract transport
layer for data exchange.

This module allows the server to initialize the client connection by sending a
`ServerInfo` message, typically used to indicate a set of supported Commands.
Client creation may fail with an error if the received `ServerInfo` does not
match the expected one.

Additionally, the `client` package includes two helper delegates:

- **KeepaliveDelegate** initiates a ping-pong exchange with the server when no
  Commands are pending. It sends the `Ping` Command and expects the `Pong`
  Result, both transmitted as a single zero byte (like a ball).
- **ReconnectDelegate** implements the `core.ClientReconnectDelegate` interface,
  providing a `Reconnect` method that should be invoked by the client if the
  connection to the server is lost.

