package session

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"irc-client/internal/app"
	"irc-client/parser"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	assert.Equal(t, "testnick", manager.GetCurrentNick())
	assert.Equal(t, "", manager.GetCurrentChannel())
	assert.False(t, manager.IsConnected())
	
	state := manager.GetClientState()
	assert.Equal(t, "testnick", state.Nick)
	assert.Equal(t, "irc.example.com", state.Server)
	assert.Equal(t, 6667, state.Port)
	assert.False(t, state.Connected)
}

func TestNickManagement(t *testing.T) {
	manager := NewManager("oldnick", "irc.example.com", 6667)
	
	// Test initial nick
	assert.Equal(t, "oldnick", manager.GetCurrentNick())
	
	// Test nick change
	manager.SetCurrentNick("newnick")
	assert.Equal(t, "newnick", manager.GetCurrentNick())
	
	// Test state reflects change
	state := manager.GetClientState()
	assert.Equal(t, "newnick", state.Nick)
}

func TestChannelManagement(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// Test initial state (no channel)
	assert.Equal(t, "", manager.GetCurrentChannel())
	
	// Test joining channel
	manager.JoinChannel("#testchan")
	assert.Equal(t, "#testchan", manager.GetCurrentChannel())
	
	// Test state reflects change
	state := manager.GetClientState()
	assert.Equal(t, "#testchan", state.CurrentChannel)
	
	// Test parting channel
	partedChannel := manager.PartChannel()
	assert.Equal(t, "#testchan", partedChannel)
	assert.Equal(t, "", manager.GetCurrentChannel())
}

func TestConnectionState(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// Test initial disconnected state
	assert.False(t, manager.IsConnected())
	
	// Test setting connected
	manager.SetConnected(true)
	assert.True(t, manager.IsConnected())
	
	// Test state reflects change
	state := manager.GetClientState()
	assert.True(t, state.Connected)
	
	// Test disconnecting
	manager.SetConnected(false)
	assert.False(t, manager.IsConnected())
}

func TestHandleMessage_Welcome(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// RPL_WELCOME should set connected state
	msg := &parser.Message{
		Command: "001",
		Params:  []string{"testnick", "Welcome to the network"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.True(t, manager.IsConnected())
}

func TestHandleMessage_Error(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	manager.SetConnected(true)
	
	// ERROR should set disconnected state
	msg := &parser.Message{
		Command: "ERROR",
		Params:  []string{"Connection lost"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.False(t, manager.IsConnected())
}

func TestHandleMessage_Join_Self(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// JOIN message where we join a channel
	msg := &parser.Message{
		Command: "JOIN",
		Prefix:  "testnick!user@host.com",
		Params:  []string{"#testchan"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "#testchan", manager.GetCurrentChannel())
}

func TestHandleMessage_Join_Other(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// JOIN message where someone else joins
	msg := &parser.Message{
		Command: "JOIN",
		Prefix:  "othernick!user@host.com",
		Params:  []string{"#testchan"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "", manager.GetCurrentChannel()) // Should not change our state
}

func TestHandleMessage_Part_Self(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	manager.JoinChannel("#testchan")
	
	// PART message where we leave the channel
	msg := &parser.Message{
		Command: "PART",
		Prefix:  "testnick!user@host.com",
		Params:  []string{"#testchan", "Goodbye"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "", manager.GetCurrentChannel())
}

func TestHandleMessage_Part_Other(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	manager.JoinChannel("#testchan")
	
	// PART message where someone else leaves
	msg := &parser.Message{
		Command: "PART",
		Prefix:  "othernick!user@host.com",
		Params:  []string{"#testchan"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "#testchan", manager.GetCurrentChannel()) // Should not change our state
}

func TestHandleMessage_Nick_Self(t *testing.T) {
	manager := NewManager("oldnick", "irc.example.com", 6667)
	
	// NICK message where we change our nick
	msg := &parser.Message{
		Command: "NICK",
		Prefix:  "oldnick!user@host.com",
		Params:  []string{"newnick"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "newnick", manager.GetCurrentNick())
}

func TestHandleMessage_Nick_Other(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// NICK message where someone else changes their nick
	msg := &parser.Message{
		Command: "NICK",
		Prefix:  "othernick!user@host.com",
		Params:  []string{"newothernick"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "testnick", manager.GetCurrentNick()) // Should not change our nick
}

func TestHandleMessage_Kick_Self(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	manager.JoinChannel("#testchan")
	
	// KICK message where we get kicked
	msg := &parser.Message{
		Command: "KICK",
		Prefix:  "operator!user@host.com",
		Params:  []string{"#testchan", "testnick", "Bad behavior"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "", manager.GetCurrentChannel())
}

func TestHandleMessage_Kick_Other(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	manager.JoinChannel("#testchan")
	
	// KICK message where someone else gets kicked
	msg := &parser.Message{
		Command: "KICK",
		Prefix:  "operator!user@host.com",
		Params:  []string{"#testchan", "othernick", "Bad behavior"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "#testchan", manager.GetCurrentChannel()) // Should not change our state
}

func TestHandleMessage_InvalidMessages(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	tests := []struct {
		name    string
		msg     *parser.Message
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "JOIN without channel",
			msg: &parser.Message{
				Command: "JOIN",
				Prefix:  "testnick!user@host.com",
				Params:  []string{},
			},
			wantErr: true,
		},
		{
			name: "PART without channel",
			msg: &parser.Message{
				Command: "PART",
				Prefix:  "testnick!user@host.com",
				Params:  []string{},
			},
			wantErr: true,
		},
		{
			name: "NICK without new nick",
			msg: &parser.Message{
				Command: "NICK",
				Prefix:  "testnick!user@host.com",
				Params:  []string{},
			},
			wantErr: true,
		},
		{
			name: "KICK without enough params",
			msg: &parser.Message{
				Command: "KICK",
				Prefix:  "operator!user@host.com",
				Params:  []string{"#testchan"}, // Missing target
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.HandleMessage(tt.msg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleMessage_UnknownCommand(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// Unknown command should not error
	msg := &parser.Message{
		Command: "UNKNOWN",
		Params:  []string{"some", "params"},
	}
	
	err := manager.HandleMessage(msg)
	assert.NoError(t, err)
}

func TestExtractNickFromPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "full prefix",
			prefix:   "nick!user@host.com",
			expected: "nick",
		},
		{
			name:     "nick only",
			prefix:   "nick",
			expected: "nick",
		},
		{
			name:     "empty prefix",
			prefix:   "",
			expected: "",
		},
		{
			name:     "prefix with multiple exclamations",
			prefix:   "nick!user!extra@host.com",
			expected: "nick",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNickFromPrefix(tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateServer(t *testing.T) {
	manager := NewManager("testnick", "old.server.com", 6667)
	
	// Test initial state
	state := manager.GetClientState()
	assert.Equal(t, "old.server.com", state.Server)
	assert.Equal(t, 6667, state.Port)
	
	// Update server
	manager.UpdateServer("new.server.com", 6697)
	
	// Test updated state
	state = manager.GetClientState()
	assert.Equal(t, "new.server.com", state.Server)
	assert.Equal(t, 6697, state.Port)
}

// Test that Manager implements app.SessionManager interface
func TestManagerImplementsInterface(t *testing.T) {
	var _ app.SessionManager = (*Manager)(nil)
}

// Concurrent access test
func TestConcurrentAccess(t *testing.T) {
	manager := NewManager("testnick", "irc.example.com", 6667)
	
	// Start goroutines that read/write concurrently
	done := make(chan bool, 4)
	
	// Goroutine 1: Change nick
	go func() {
		for i := 0; i < 100; i++ {
			manager.SetCurrentNick("nick1")
			manager.SetCurrentNick("nick2")
		}
		done <- true
	}()
	
	// Goroutine 2: Join/part channels
	go func() {
		for i := 0; i < 100; i++ {
			manager.JoinChannel("#chan1")
			manager.PartChannel()
		}
		done <- true
	}()
	
	// Goroutine 3: Read state
	go func() {
		for i := 0; i < 100; i++ {
			manager.GetCurrentNick()
			manager.GetCurrentChannel()
			manager.GetClientState()
		}
		done <- true
	}()
	
	// Goroutine 4: Connection state
	go func() {
		for i := 0; i < 100; i++ {
			manager.SetConnected(true)
			manager.SetConnected(false)
		}
		done <- true
	}()
	
	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}
	
	// Test should complete without data races
	assert.True(t, true)
}