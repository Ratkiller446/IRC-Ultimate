package commands

import (
        "strings"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"

        "irc-client/internal/app"
)

// Test backward compatibility with existing commands_test.go tests
func TestBuildIRCCommand_BackwardCompatibility(t *testing.T) {
        // Test some key cases from the original tests to ensure compatibility
        tests := []struct {
                name      string
                state     ClientState
                userInput string
                expected  []string
                expectErr bool
        }{
                {
                        name:      "join with channel",
                        state:     ClientState{CurrentChannel: "", Nick: "testuser"},
                        userInput: "/join #testchannel",
                        expected:  []string{"JOIN #testchannel"},
                        expectErr: false,
                },
                {
                        name:      "join with channel and key",
                        state:     ClientState{CurrentChannel: "", Nick: "testuser"},
                        userInput: "/join #private secretkey",
                        expected:  []string{"JOIN #private secretkey"},
                        expectErr: false,
                },
                {
                        name:      "part from current channel",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testuser"},
                        userInput: "/part",
                        expected:  []string{"PART #testchannel"},
                        expectErr: false,
                },
                {
                        name:      "part with reason",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testuser"},
                        userInput: "/part Goodbye everyone!",
                        expected:  []string{"PART #testchannel :Goodbye everyone!"},
                        expectErr: false,
                },
                {
                        name:      "nick change",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "oldnick"},
                        userInput: "/nick newnick",
                        expected:  []string{"NICK newnick"},
                        expectErr: false,
                },
                {
                        name:      "msg to user",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg targetuser Hello there!",
                        expected:  []string{"PRIVMSG targetuser :Hello there!"},
                        expectErr: false,
                },
                {
                        name:      "quit without reason",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/quit",
                        expected:  []string{"QUIT"},
                        expectErr: false,
                },
                {
                        name:      "quit with reason",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/quit Goodbye everyone!",
                        expected:  []string{"QUIT :Goodbye everyone!"},
                        expectErr: false,
                },
                {
                        name:      "channel message",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "Hello everyone!",
                        expected:  []string{"PRIVMSG #testchannel :Hello everyone!"},
                        expectErr: false,
                },
                {
                        name:      "channel message with no current channel",
                        state:     ClientState{CurrentChannel: "", Nick: "testnick"},
                        userInput: "Hello everyone!",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "unknown command",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/unknown some args",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "empty input",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "just slash",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/",
                        expected:  []string{},
                        expectErr: false,
                },
        }
        
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        result, err := BuildIRCCommand(tt.state, tt.userInput)
                        
                        if tt.expectErr {
                                assert.Error(t, err)
                        } else {
                                require.NoError(t, err)
                                assert.Equal(t, tt.expected, result)
                        }
                })
        }
}

func TestBuildIRCCommand_Sanitization(t *testing.T) {
        state := ClientState{CurrentChannel: "#test", Nick: "testnick"}
        
        // Test that sanitization still works through the legacy interface
        result, err := BuildIRCCommand(state, "/join #test\r\n\x00channel")
        require.NoError(t, err)
        require.Len(t, result, 1)
        assert.Equal(t, "JOIN #testchannel", result[0])
}

func TestBuildIRCCommand_NewFeatures(t *testing.T) {
        // Test that the new router has additional capabilities
        // while maintaining backward compatibility
        state := ClientState{CurrentChannel: "#test", Nick: "testnick"}
        
        // The legacy function should still work even if the router has new handlers
        router := NewRouter()
        
        // Add a new handler to test extensibility
        err := router.RegisterHandler("test", &TestLegacyHandler{})
        require.NoError(t, err)
        
        // Legacy function should still work for existing commands
        result, err := BuildIRCCommand(state, "/join #channel")
        require.NoError(t, err)
        assert.Equal(t, []string{"JOIN #channel"}, result)
        
        // But new commands won't be available through legacy interface
        // (which is expected behavior for backward compatibility)
        result, err = BuildIRCCommand(state, "/test args")
        require.NoError(t, err)
        assert.Empty(t, result) // Unknown command returns empty
}

// TestLegacyHandler for testing extensibility
type TestLegacyHandler struct{}

func (h *TestLegacyHandler) Handle(state app.ClientState, args []string) ([]string, error) {
        return []string{"TEST " + strings.Join(args, " ")}, nil
}

func (h *TestLegacyHandler) Description() string {
        return "Test legacy handler"
}

func TestClientStateConversion(t *testing.T) {
        // Test that legacy ClientState properly converts to app.ClientState
        legacyState := ClientState{
                CurrentChannel: "#test",
                Nick:           "testnick",
        }
        
        // This conversion happens inside BuildIRCCommand
        result, err := BuildIRCCommand(legacyState, "/join #newchan")
        require.NoError(t, err)
        assert.Equal(t, []string{"JOIN #newchan"}, result)
        
        // Test edge cases
        emptyState := ClientState{}
        result, err = BuildIRCCommand(emptyState, "hello")
        require.NoError(t, err)
        assert.Empty(t, result) // No current channel, so message is dropped
}