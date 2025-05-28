# delegate-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/delegate-go.svg)](https://pkg.go.dev/github.com/cmd-stream/delegate-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/base-go)](https://goreportcard.com/report/github.com/cmd-stream/base-go)
[![codecov](https://codecov.io/gh/cmd-stream/delegate-go/graph/badge.svg?token=G8NN40DYJI)](https://codecov.io/gh/cmd-stream/delegate-go)

`delegate-go` facilitates communication between the cmd-stream client and server.
It provides implementations of the `base.ClientDelegate` and `base.ServerDelegate` 
interfaces, both of which rely on an abstract transport layer for data exchange.

This module allows the server to initialize the client connection by sending a 
`ServerInfo` byte slice, typically used to indicate a set of supported commands. 
Client creation may fail with an error if it's `ServerInfo` doen't match the one
received from the server.

Additionally, the `client` package includes `KeepaliveDelegate` and 
`ReconnectDelegate`:
- `KeepaliveDelegate` initiates a ping-pong exchange with the server when no 
  commands are pending. It sends the `Ping` command and expects the `Pong` 
  result, both transmitted as a single zero byte (like a ball).
- `ReconnectDelegate` implements the `base.ClientReconnectDelegate` interface, 
  providing a `Reconnect` method that should be invoked by the client if the
  connection to the server is lost.