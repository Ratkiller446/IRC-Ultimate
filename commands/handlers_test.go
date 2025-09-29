package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"irc-client/internal/app"
)

func TestJoinHandler(t *testing.T) {
	handler := &JoinHandler{}
	state := app.ClientState{Nick: "testnick"}
	
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "join with channel",
			args:     []string{"#test"},
			expected: []string{"JOIN #test"},
		},
		{
			name:     "join with channel and key",
			args:     []string{"#private", "secret"},
			expected: []string{"JOIN #private secret"},
		},
		{
			name:     "join with no args",
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "Join a channel. Usage: /join #channel [key]", handler.Description())
}

func TestPartHandler(t *testing.T) {
	handler := &PartHandler{}
	
	tests := []struct {
		name     string
		state    app.ClientState
		args     []string
		expected []string
	}{
		{
			name:     "part without reason",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{},
			expected: []string{"PART #test"},
		},
		{
			name:     "part with reason",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{"goodbye", "everyone"},
			expected: []string{"PART #test :goodbye everyone"},
		},
		{
			name:     "part with no current channel",
			state:    app.ClientState{CurrentChannel: "", Nick: "testnick"},
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(tt.state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "Leave the current channel. Usage: /part [reason]", handler.Description())
}

func TestNickHandler(t *testing.T) {
	handler := &NickHandler{}
	state := app.ClientState{Nick: "oldnick"}
	
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "nick change",
			args:     []string{"newnick"},
			expected: []string{"NICK newnick"},
		},
		{
			name:     "nick with no args",
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "Change your nickname. Usage: /nick newnick", handler.Description())
}

func TestMsgHandler(t *testing.T) {
	handler := &MsgHandler{}
	state := app.ClientState{Nick: "testnick"}
	
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "msg to user",
			args:     []string{"target", "hello", "world"},
			expected: []string{"PRIVMSG target :hello world"},
		},
		{
			name:     "msg with just target",
			args:     []string{"target"},
			expected: []string{},
		},
		{
			name:     "msg with no args",
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "Send a private message. Usage: /msg target message", handler.Description())
}

func TestQuitHandler(t *testing.T) {
	handler := &QuitHandler{}
	state := app.ClientState{Nick: "testnick"}
	
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "quit without reason",
			args:     []string{},
			expected: []string{"QUIT"},
		},
		{
			name:     "quit with reason",
			args:     []string{"goodbye", "cruel", "world"},
			expected: []string{"QUIT :goodbye cruel world"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "Quit the IRC server. Usage: /quit [reason]", handler.Description())
}

func TestHelpHandler(t *testing.T) {
	router := NewRouter()
	handler := NewHelpHandler(router)
	state := app.ClientState{Nick: "testnick"}
	
	// Help returns empty commands (displayed locally)
	result, err := handler.Handle(state, []string{})
	require.NoError(t, err)
	assert.Empty(t, result)
	
	// Test general help
	helpText := handler.GetHelpText("")
	assert.Contains(t, helpText, "Available commands:")
	assert.Contains(t, helpText, "/join")
	assert.Contains(t, helpText, "/part")
	
	// Test specific command help
	helpText = handler.GetHelpText("join")
	assert.Equal(t, "Join a channel. Usage: /join #channel [key]", helpText)
	
	// Test unknown command help
	helpText = handler.GetHelpText("unknown")
	assert.Equal(t, "Unknown command: unknown", helpText)
	
	assert.Equal(t, "Show help for commands. Usage: /help [command]", handler.Description())
}

func TestWhoHandler(t *testing.T) {
	handler := &WhoHandler{}
	
	tests := []struct {
		name     string
		state    app.ClientState
		args     []string
		expected []string
	}{
		{
			name:     "who current channel",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{},
			expected: []string{"WHO #test"},
		},
		{
			name:     "who specific target",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{"#other"},
			expected: []string{"WHO #other"},
		},
		{
			name:     "who with no current channel",
			state:    app.ClientState{CurrentChannel: "", Nick: "testnick"},
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(tt.state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "List users in a channel. Usage: /who [channel]", handler.Description())
}

func TestTopicHandler(t *testing.T) {
	handler := &TopicHandler{}
	
	tests := []struct {
		name     string
		state    app.ClientState
		args     []string
		expected []string
	}{
		{
			name:     "get topic",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{},
			expected: []string{"TOPIC #test"},
		},
		{
			name:     "set topic",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{"New", "topic", "here"},
			expected: []string{"TOPIC #test :New topic here"},
		},
		{
			name:     "topic with no current channel",
			state:    app.ClientState{CurrentChannel: "", Nick: "testnick"},
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(tt.state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "View or set channel topic. Usage: /topic [new topic]", handler.Description())
}

func TestMeHandler(t *testing.T) {
	handler := &MeHandler{}
	
	tests := []struct {
		name     string
		state    app.ClientState
		args     []string
		expected []string
	}{
		{
			name:     "action message",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{"waves", "hello"},
			expected: []string{"PRIVMSG #test :\x01ACTION waves hello\x01"},
		},
		{
			name:     "me with no current channel",
			state:    app.ClientState{CurrentChannel: "", Nick: "testnick"},
			args:     []string{"waves"},
			expected: []string{},
		},
		{
			name:     "me with no args",
			state:    app.ClientState{CurrentChannel: "#test", Nick: "testnick"},
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Handle(tt.state, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
	
	assert.Equal(t, "Send an action message. Usage: /me action", handler.Description())
}

// Test that all handlers implement app.CommandHandler interface
func TestHandlers_ImplementInterface(t *testing.T) {
	var _ app.CommandHandler = (*JoinHandler)(nil)
	var _ app.CommandHandler = (*PartHandler)(nil)
	var _ app.CommandHandler = (*NickHandler)(nil)
	var _ app.CommandHandler = (*MsgHandler)(nil)
	var _ app.CommandHandler = (*QuitHandler)(nil)
	var _ app.CommandHandler = (*HelpHandler)(nil)
	var _ app.CommandHandler = (*WhoHandler)(nil)
	var _ app.CommandHandler = (*TopicHandler)(nil)
	var _ app.CommandHandler = (*MeHandler)(nil)
}