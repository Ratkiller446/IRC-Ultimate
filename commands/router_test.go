package commands

import (
        "strings"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"

        "irc-client/internal/app"
)

func TestNewRouter(t *testing.T) {
        router := NewRouter()
        
        // Should have default handlers
        expectedHandlers := []string{"join", "part", "nick", "msg", "quit"}
        for _, handler := range expectedHandlers {
                assert.True(t, router.HasHandler(handler), "Router should have %s handler", handler)
        }
        
        // Should be able to get registered commands
        commands := router.GetRegisteredCommands()
        assert.Len(t, commands, 5)
}

func TestRouter_Route_SlashCommands(t *testing.T) {
        router := NewRouter()
        state := app.ClientState{
                CurrentChannel: "#test",
                Nick:           "testnick",
                Connected:      true,
        }
        
        tests := []struct {
                name           string
                input          string
                expectedCount  int
                expectedType   app.CommandType
                expectedTarget string
                expectedRaw    string
        }{
                {
                        name:           "join command",
                        input:          "/join #channel",
                        expectedCount:  1,
                        expectedType:   app.CommandTypeJoin,
                        expectedTarget: "#channel",
                        expectedRaw:    "JOIN #channel",
                },
                {
                        name:           "join with key",
                        input:          "/join #private secret",
                        expectedCount:  1,
                        expectedType:   app.CommandTypeJoin,
                        expectedTarget: "#private",
                        expectedRaw:    "JOIN #private secret",
                },
                {
                        name:           "part command",
                        input:          "/part",
                        expectedCount:  1,
                        expectedType:   app.CommandTypePart,
                        expectedTarget: "#test",
                        expectedRaw:    "PART #test",
                },
                {
                        name:           "part with reason",
                        input:          "/part goodbye",
                        expectedCount:  1,
                        expectedType:   app.CommandTypePart,
                        expectedTarget: "#test",
                        expectedRaw:    "PART #test :goodbye",
                },
                {
                        name:           "nick command",
                        input:          "/nick newnick",
                        expectedCount:  1,
                        expectedType:   app.CommandTypeNick,
                        expectedTarget: "newnick",
                        expectedRaw:    "NICK newnick",
                },
                {
                        name:           "msg command",
                        input:          "/msg target hello",
                        expectedCount:  1,
                        expectedType:   app.CommandTypePrivMsg,
                        expectedTarget: "target",
                        expectedRaw:    "PRIVMSG target :hello",
                },
                {
                        name:           "quit command",
                        input:          "/quit",
                        expectedCount:  1,
                        expectedType:   app.CommandTypeQuit,
                        expectedRaw:    "QUIT",
                },
                {
                        name:           "quit with reason",
                        input:          "/quit goodbye",
                        expectedCount:  1,
                        expectedType:   app.CommandTypeQuit,
                        expectedRaw:    "QUIT :goodbye",
                },
                {
                        name:          "unknown command",
                        input:         "/unknown args",
                        expectedCount: 0,
                },
                {
                        name:          "just slash",
                        input:         "/",
                        expectedCount: 0,
                },
                {
                        name:          "empty command",
                        input:         "/join",
                        expectedCount: 0,
                },
        }
        
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        commands, err := router.Route(state, tt.input)
                        require.NoError(t, err)
                        assert.Len(t, commands, tt.expectedCount)
                        
                        if tt.expectedCount > 0 {
                                cmd := commands[0]
                                assert.Equal(t, tt.expectedType, cmd.Type)
                                assert.Equal(t, tt.expectedRaw, cmd.Raw)
                                if tt.expectedTarget != "" {
                                        assert.Equal(t, tt.expectedTarget, cmd.Target)
                                }
                        }
                })
        }
}

func TestRouter_Route_ChannelMessages(t *testing.T) {
        router := NewRouter()
        
        tests := []struct {
                name          string
                state         app.ClientState
                input         string
                expectedCount int
                expectedRaw   string
        }{
                {
                        name: "channel message",
                        state: app.ClientState{
                                CurrentChannel: "#test",
                                Nick:           "testnick",
                        },
                        input:         "hello world",
                        expectedCount: 1,
                        expectedRaw:   "PRIVMSG #test :hello world",
                },
                {
                        name: "no current channel",
                        state: app.ClientState{
                                CurrentChannel: "",
                                Nick:           "testnick",
                        },
                        input:         "hello world",
                        expectedCount: 0,
                },
                {
                        name: "empty message",
                        state: app.ClientState{
                                CurrentChannel: "#test",
                                Nick:           "testnick",
                        },
                        input:         "",
                        expectedCount: 0,
                },
        }
        
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        commands, err := router.Route(tt.state, tt.input)
                        require.NoError(t, err)
                        assert.Len(t, commands, tt.expectedCount)
                        
                        if tt.expectedCount > 0 {
                                cmd := commands[0]
                                assert.Equal(t, app.CommandTypePrivMsg, cmd.Type)
                                assert.Equal(t, tt.expectedRaw, cmd.Raw)
                                assert.Equal(t, tt.state.CurrentChannel, cmd.Target)
                                assert.Equal(t, tt.input, cmd.Message)
                        }
                })
        }
}

func TestRouter_RegisterHandler(t *testing.T) {
        router := NewRouter()
        
        // Test successful registration
        testHandler := &TestHandler{}
        err := router.RegisterHandler("test", testHandler)
        assert.NoError(t, err)
        assert.True(t, router.HasHandler("test"))
        
        // Test case insensitive
        assert.True(t, router.HasHandler("TEST"))
        
        // Test error cases
        err = router.RegisterHandler("", testHandler)
        assert.Error(t, err)
        
        err = router.RegisterHandler("test2", nil)
        assert.Error(t, err)
        
        // Test override
        newHandler := &TestHandler{}
        err = router.RegisterHandler("test", newHandler)
        assert.NoError(t, err)
}

func TestRouter_UnregisterHandler(t *testing.T) {
        router := NewRouter()
        
        // Test unregistering existing handler
        assert.True(t, router.HasHandler("join"))
        err := router.UnregisterHandler("join")
        assert.NoError(t, err)
        assert.False(t, router.HasHandler("join"))
        
        // Test error cases
        err = router.UnregisterHandler("")
        assert.Error(t, err)
        
        err = router.UnregisterHandler("nonexistent")
        assert.Error(t, err)
}

func TestRouter_StateChanges(t *testing.T) {
        router := NewRouter()
        state := app.ClientState{
                CurrentChannel: "#test",
                Nick:           "testnick",
        }
        
        tests := []struct {
                name             string
                input            string
                expectedChange   bool
                channelAction    app.ChannelAction
                newChannel       string
                newNick          string
                shouldQuit       bool
        }{
                {
                        name:           "join creates state change",
                        input:          "/join #newchan",
                        expectedChange: true,
                        channelAction:  app.ChannelActionJoin,
                        newChannel:     "#newchan",
                },
                {
                        name:           "part creates state change",
                        input:          "/part",
                        expectedChange: true,
                        channelAction:  app.ChannelActionPart,
                        newChannel:     "#test",
                },
                {
                        name:           "nick creates state change",
                        input:          "/nick newnick",
                        expectedChange: true,
                        newNick:        "newnick",
                },
                {
                        name:           "quit creates state change",
                        input:          "/quit",
                        expectedChange: true,
                        shouldQuit:     true,
                },
                {
                        name:           "msg no state change",
                        input:          "/msg target hello",
                        expectedChange: false,
                },
        }
        
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        commands, err := router.Route(state, tt.input)
                        require.NoError(t, err)
                        require.Len(t, commands, 1)
                        
                        cmd := commands[0]
                        if tt.expectedChange {
                                require.NotNil(t, cmd.StateChange)
                                
                                if tt.newChannel != "" {
                                        require.NotNil(t, cmd.StateChange.ChannelChange)
                                        assert.Equal(t, tt.channelAction, cmd.StateChange.ChannelChange.Action)
                                        assert.Equal(t, tt.newChannel, cmd.StateChange.ChannelChange.Channel)
                                }
                                
                                if tt.newNick != "" {
                                        require.NotNil(t, cmd.StateChange.NickChange)
                                        assert.Equal(t, tt.newNick, cmd.StateChange.NickChange.NewNick)
                                }
                                
                                if tt.shouldQuit {
                                        assert.True(t, cmd.StateChange.ShouldQuit)
                                }
                        } else {
                                assert.Nil(t, cmd.StateChange)
                        }
                })
        }
}

func TestRouter_Sanitization(t *testing.T) {
        router := NewRouter()
        state := app.ClientState{
                CurrentChannel: "#test",
                Nick:           "testnick",
        }
        
        // Test that input is sanitized
        commands, err := router.Route(state, "/join #test\r\n\x00channel")
        require.NoError(t, err)
        require.Len(t, commands, 1)
        
        // Should have cleaned input
        assert.Equal(t, "JOIN #testchannel", commands[0].Raw)
}

// TestHandler is a mock handler for testing
type TestHandler struct {
        handleCalled bool
        lastState    app.ClientState
        lastArgs     []string
}

func (h *TestHandler) Handle(state app.ClientState, args []string) ([]string, error) {
        h.handleCalled = true
        h.lastState = state
        h.lastArgs = args
        return []string{"TEST " + strings.Join(args, " ")}, nil
}

func (h *TestHandler) Description() string {
        return "Test handler"
}

func TestRouter_HandlerExecution(t *testing.T) {
        router := NewRouter()
        testHandler := &TestHandler{}
        
        err := router.RegisterHandler("test", testHandler)
        require.NoError(t, err)
        
        state := app.ClientState{
                CurrentChannel: "#test",
                Nick:           "testnick",
        }
        
        commands, err := router.Route(state, "/test arg1 arg2")
        require.NoError(t, err)
        require.Len(t, commands, 1)
        
        // Verify handler was called
        assert.True(t, testHandler.handleCalled)
        assert.Equal(t, state, testHandler.lastState)
        assert.Equal(t, []string{"arg1", "arg2"}, testHandler.lastArgs)
        assert.Equal(t, "TEST arg1 arg2", commands[0].Raw)
}

// Test that Router implements app.CommandRouter interface
func TestRouter_ImplementsInterface(t *testing.T) {
        var _ app.CommandRouter = (*Router)(nil)
}