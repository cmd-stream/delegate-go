// Package server provides a server-side implementation for the delegate
// abstraction of the cmd-stream library.
//
// It defines the Delegate type, which implements the server.Delegate interface
// Defined in the core-go module. Delegate sends ServerInfo to initialize the
// client connection and then handles Commands via a TransportHandler.
//
// Deprecated: migrate to github.com/cmd-stream/cmd-stream-go instead.
package server
