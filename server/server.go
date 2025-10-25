// Package server provides server-side implementations for the delegate
// abstraction of the cmd-stream-go library.
//
// It defines the Delegate type, which implements the core.ServerDelegate
// interface. Delegate sends ServerInfo to initialize the client connection
// and then handles Commands via a TransportHandler.
package server
