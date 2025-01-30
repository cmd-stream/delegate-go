# delegate-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/delegate-go.svg)](https://pkg.go.dev/github.com/cmd-stream/delegate-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/base-go)](https://goreportcard.com/report/github.com/cmd-stream/base-go)
[![codecov](https://codecov.io/gh/cmd-stream/delegate-go/graph/badge.svg?token=G8NN40DYJI)](https://codecov.io/gh/cmd-stream/delegate-go)

delegate-go provides communication between the cmd-stream client and server.

It includes implementations of the `base.ClientDelegate` and `base.ServerDelegate` 
interfaces, each of which relies on an abstract transport for data delivery.

This module allows the server to initialize the client connection by sending
`ServerInfo` - a slice of bytes, that may denote, for example a set of supported 
Commands. If the client's `ServerInfo` does not match the one received from the 
server, client creation will fail with an error.

Also in the `client` package you can find `KeepaliveDelegate` and 
`ReconnectDelegate`. `KeepaliveDelegate` starts playing the Ping-Pong game with 
the server when there are no Commands to send - it sends the `Ping` Command and 
receives the `Pong` result, both of which are transmitted as a 0 (like a ball) 
byte. `ReconnectDelegate` is an implementation of the `base.ClientReconnectDelegate`
interface, it has a `Reconnect` method that should be used by the client if the 
connection to the server has been lost.