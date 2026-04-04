// Package client provides client-side implementations for the delegate
// abstraction of the cmd-stream library.
//
// It defines several delegate types that implement the client.Delegate
// client.KeepaliveDelegate and client.ReconnectDelegate interfaces defined in
// the core-go module.
//
//   - Delegate: Basic client delegate that receives ServerInfo.
//   - KeepaliveDelegate: Extends Delegate with a ping-pong mechanism to keep
//     the connection alive when no Commands are pending.
//   - ReconnectDelegate: Extends Delegate with automatic reconnect logic
//     when the connection to the server is lost.
//
// All delegates rely on a pluggable Transport for data exchange and support
// configurable options such as send/receive deadlines.
//
// Deprecated: migrate to github.com/cmd-stream/cmd-stream-go instead.
package client
