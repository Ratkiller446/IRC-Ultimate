package conn

import (
	"bufio"
	"context"

	"irc-client/internal/app"
)

// Adapter wraps Manager to implement app.Connection interface
type Adapter struct {
	manager *Manager
}

// NewAdapter creates a new connection adapter that wraps the existing Manager
func NewAdapter(config Config) app.Connection {
	manager := NewManager(config)
	return &Adapter{
		manager: manager,
	}
}

// NewAdapterFromManager creates an adapter from an existing Manager
func NewAdapterFromManager(manager *Manager) app.Connection {
	return &Adapter{
		manager: manager,
	}
}

// Connect establishes connection to IRC server
func (a *Adapter) Connect() error {
	return a.manager.Connect()
}

// Write sends data to the IRC server
func (a *Adapter) Write(data string) error {
	return a.manager.Write(data)
}

// Reader returns a reader for incoming data
func (a *Adapter) Reader() *bufio.Reader {
	return a.manager.Reader()
}

// ErrorChannel returns channel for connection errors
func (a *Adapter) ErrorChannel() <-chan error {
	return a.manager.ErrorChannel()
}

// Close gracefully shuts down the connection
func (a *Adapter) Close() error {
	return a.manager.Close()
}

// Context returns context for coordinated shutdown
func (a *Adapter) Context() context.Context {
	return a.manager.Context()
}

// IsConnected returns true if connection is established
func (a *Adapter) IsConnected() bool {
	return a.manager.IsConnected()
}

// GetManager returns the underlying Manager for advanced use cases
// This allows access to Manager-specific methods while still using the interface
func (a *Adapter) GetManager() *Manager {
	return a.manager
}

// GetState returns the current connection state
func (a *Adapter) GetState() ConnectionState {
	return a.manager.GetState()
}