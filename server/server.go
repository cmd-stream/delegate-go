// Package server provides a server-side implementation for the delegate
// abstraction of the cmd-stream library.
//
// It defines the Delegate type, which implements the server.Delegate interface
// defined in the core-go module. Delegate sends ServerInfo to initialize the
// client connection and then handles Commands via a TransportHandler.
package server
