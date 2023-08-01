# delegate-go
delegate-go provides communication between the cmd-stream client and server for 
Golang.

It contains implementations of the `base.ClientDelegate` and 
`base.ServerDelegate` interfaces (they are located in the corresponding 
packages).

A feature of this module is the system data that the server sends to the client.
First of all it sends the `ServerInfo`, than `ServerSettings`.

With the `ServerInfo` you can define, for example, cmdspace, which can identify 
a set of commands supported by the server. If the `ServerInfo` of the client 
does not match with the `ServerInfo` received from the server, the client 
creation will end in an error.

With the `ServerSettings` you can define maximum command size supported by the 
server. The client should not send commands of a larger size.

In turn, each of the delegates depends on an abstract transport that delivers 
data from the client to the server and vice versa.

Also in the `client` package you can find a `KeepaliveDelegate` and 
`ReconnectDelegate`.

The `KeepaliveDelegate` starts playing the Ping-Pong game with the server when 
there are no commands to send - it sends a `Ping` command and receives a `Pong` 
result, both of which are transmitted as a 0 (like a ball) byte.

The `ReconnectDelegate` is an implementation of the `base.ClientReconnectDelegate`
interface, it has a `Reconnect` method that can be used by the client if the 
connection to the server has been lost.

# Tests
Test coverage is about 93%.