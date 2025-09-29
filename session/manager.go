package session

import (
	"fmt"
	"strings"
	"sync"

	"irc-client/internal/app"
	"irc-client/parser"
)

// Manager implements app.SessionManager interface
type Manager struct {
	mu             sync.RWMutex
	currentNick    string
	currentChannel string
	connected      bool
	server         string
	port           int
}

// NewManager creates a new session manager
func NewManager(nick, server string, port int) *Manager {
	return &Manager{
		currentNick: nick,
		server:      server,
		port:        port,
		connected:   false,
	}
}

// GetCurrentNick returns the current nickname
func (m *Manager) GetCurrentNick() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentNick
}

// SetCurrentNick updates the current nickname
func (m *Manager) SetCurrentNick(nick string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentNick = nick
}

// GetCurrentChannel returns the currently active channel
func (m *Manager) GetCurrentChannel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentChannel
}

// JoinChannel updates state when joining a channel
func (m *Manager) JoinChannel(channel string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentChannel = channel
}

// PartChannel updates state when leaving a channel and returns the channel that was left
func (m *Manager) PartChannel() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	previousChannel := m.currentChannel
	m.currentChannel = ""
	return previousChannel
}

// SetConnected updates the connection state
func (m *Manager) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

// IsConnected returns true if connected to server
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// HandleMessage processes incoming IRC messages and updates session state
func (m *Manager) HandleMessage(msg *parser.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// Handle different message types that affect session state
	switch msg.Command {
	case "JOIN":
		return m.handleJoin(msg)
	case "PART":
		return m.handlePart(msg)
	case "NICK":
		return m.handleNick(msg)
	case "KICK":
		return m.handleKick(msg)
	case "001": // RPL_WELCOME - successful connection
		m.SetConnected(true)
		return nil
	case "ERROR":
		m.SetConnected(false)
		return nil
	}

	return nil // No error, just not a state-affecting message
}

// handleJoin processes JOIN messages
func (m *Manager) handleJoin(msg *parser.Message) error {
	if len(msg.Params) < 1 {
		return fmt.Errorf("JOIN message missing channel parameter")
	}

	// Extract nickname from prefix (nick!user@host)
	nick := extractNickFromPrefix(msg.Prefix)
	
	// Only update our state if we're the one joining
	if nick == m.GetCurrentNick() {
		m.JoinChannel(msg.Params[0])
	}

	return nil
}

// handlePart processes PART messages
func (m *Manager) handlePart(msg *parser.Message) error {
	if len(msg.Params) < 1 {
		return fmt.Errorf("PART message missing channel parameter")
	}

	// Extract nickname from prefix
	nick := extractNickFromPrefix(msg.Prefix)
	
	// Only update our state if we're the one parting
	if nick == m.GetCurrentNick() {
		// Check if we're parting from our current channel
		if msg.Params[0] == m.GetCurrentChannel() {
			m.PartChannel()
		}
	}

	return nil
}

// handleNick processes NICK messages
func (m *Manager) handleNick(msg *parser.Message) error {
	if len(msg.Params) < 1 {
		return fmt.Errorf("NICK message missing new nickname parameter")
	}

	// Extract old nickname from prefix
	oldNick := extractNickFromPrefix(msg.Prefix)
	newNick := msg.Params[0]
	
	// Update our nick if we're the one changing it
	if oldNick == m.GetCurrentNick() {
		m.SetCurrentNick(newNick)
	}

	return nil
}

// handleKick processes KICK messages
func (m *Manager) handleKick(msg *parser.Message) error {
	if len(msg.Params) < 2 {
		return fmt.Errorf("KICK message missing channel or target parameters")
	}

	channel := msg.Params[0]
	target := msg.Params[1]
	
	// If we were kicked, clear the channel state
	if target == m.GetCurrentNick() && channel == m.GetCurrentChannel() {
		m.PartChannel()
	}

	return nil
}

// GetClientState returns current state for command processing
func (m *Manager) GetClientState() app.ClientState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return app.ClientState{
		CurrentChannel: m.currentChannel,
		Nick:           m.currentNick,
		Connected:      m.connected,
		Server:         m.server,
		Port:           m.port,
	}
}

// extractNickFromPrefix extracts nickname from IRC prefix (nick!user@host)
func extractNickFromPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	
	// Find the first '!' to separate nick from user@host
	exclamIndex := strings.IndexByte(prefix, '!')
	if exclamIndex == -1 {
		// No '!' found, return the whole prefix as nick
		return prefix
	}
	
	return prefix[:exclamIndex]
}

// UpdateServer updates the server information
func (m *Manager) UpdateServer(server string, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.server = server
	m.port = port
}