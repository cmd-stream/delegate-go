# delegate-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/delegate-go.svg)](https://pkg.go.dev/github.com/cmd-stream/delegate-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/base-go)](https://goreportcard.com/report/github.com/cmd-stream/base-go)
[![codecov](https://codecov.io/gh/cmd-stream/delegate-go/graph/badge.svg?token=G8NN40DYJI)](https://codecov.io/gh/cmd-stream/delegate-go)

delegate-go facilitates communication between the cmd-stream client and server.

It includes implementations of the `base.ClientDelegate` and `base.ServerDelegate`
interfaces, each of which relies on an abstract transport for data delivery.

This module allows the server to initialize the client connection by sending
`ServerInfo` - a byte slice that, for example, may denote a set of supported 
Commands. If the client's `ServerInfo` does not match the one received from the 
server, client creation will fail with an error.

Additionally, the `client` package provides `KeepaliveDelegate` and 
`ReconnectDelegate`:
- `KeepaliveDelegate` initiates a Ping-Pong exchange with the server when there
  are no Commands to send - it sends the `Ping` Command and receives the `Pong` 
  result, both transmitted as a single zero byte (like a ball).
- `ReconnectDelegate` implements the `base.ClientReconnectDelegate` interface,
  providing a `Reconnect` method that the client should invoke if the connection
  to the server is lost.