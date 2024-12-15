# delegate-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/delegate-go.svg)](https://pkg.go.dev/github.com/cmd-stream/delegate-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/base-go)](https://goreportcard.com/report/github.com/cmd-stream/base-go)
[![codecov](https://codecov.io/gh/cmd-stream/delegate-go/graph/badge.svg?token=G8NN40DYJI)](https://codecov.io/gh/cmd-stream/delegate-go)

delegate-go provides communication between the cmd-stream-go client and server.

It contains implementations of the `base.ClientDelegate` and 
`base.ServerDelegate` interfaces (they are located in the corresponding 
packages). Each of the delegates depends on an abstract transport that delivers 
data from the client to the server and vice versa.

With this module, the server initializes the client connection by sending system
data: `ServerInfo` and `ServerSettings`. With `ServerInfo` you can define, for 
example, cmdspace, which identifies the set of commands supported by the server. 
If the client's `ServerInfo` does not match the `ServerInfo` received from the 
server, client creation will fail with an error. With `ServerSettings` you can 
define maximum command size supported by the server. The client should use this 
setting to validate commands before sending them.

Also in the `client` package you can find `KeepaliveDelegate` and 
`ReconnectDelegate`. `KeepaliveDelegate` starts playing the Ping-Pong game with 
the server when there are no commands to send - it sends the `Ping` command and 
receives the `Pong` result, both of which are transmitted as a 0 (like a ball) 
byte. `ReconnectDelegate` is an implementation of the `base.ClientReconnectDelegate`
interface, it has a `Reconnect` method that should be used by the client if the 
connection to the server has been lost.
