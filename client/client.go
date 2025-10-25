// Package client provides client-side implementations for the delegate
// abstraction of the cmd-stream-go library.
//
// It defines several Delegate types that implement the core.ClientDelegate
// and core.ClientReconnectDelegate interfaces.
//
// Key delegates:
//
//   - Delegate: basic client delegate that receives ServerInfo.
//   - KeepaliveDelegate: extends Delegate with a ping-pong mechanism to keep
//     the connection alive when no Commands are pending.
//   - ReconnectDelegate: extends Delegate with automatic reconnect logic
//     when the connection to the server is lost.
//
// All delegates rely on a pluggable Transport for data exchange and support
// configurable options such as send/receive deadlines.
package client
