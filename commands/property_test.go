package commands

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"irc-client/testutil"
)

// TestSanitize_PropertyBased uses property-based testing to verify input sanitization
func TestSanitize_PropertyBased(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Sanitize removes all dangerous characters
	properties.Property("sanitize removes dangerous characters", prop.ForAll(
		func(input string) bool {
			sanitized := Sanitize(input)
			
			// Verify no dangerous characters remain
			return !strings.Contains(sanitized, "\r") &&
				!strings.Contains(sanitized, "\n") &&
				!strings.Contains(sanitized, "\x00")
		},
		testutil.UserInputGenerator(),
	))

	// Property: Sanitize is idempotent
	properties.Property("sanitize is idempotent", prop.ForAll(
		func(input string) bool {
			once := Sanitize(input)
			twice := Sanitize(once)
			return once == twice
		},
		testutil.UserInputGenerator(),
	))

	// Property: Safe strings are unchanged
	properties.Property("safe strings are unchanged", prop.ForAll(
		func(safe string) bool {
			// Only test strings that don't contain dangerous characters
			if strings.Contains(safe, "\r") || 
				strings.Contains(safe, "\n") || 
				strings.Contains(safe, "\x00") {
				return true // Skip unsafe inputs
			}
			
			return Sanitize(safe) == safe
		},
		testutil.UserInputGenerator(),
	))

	// Property: Empty string handling
	properties.Property("empty string is handled correctly", prop.ForAll(
		func() bool {
			return Sanitize("") == ""
		},
	))

	properties.TestingRun(t)
}

// TestBuildIRCCommand_PropertyBased tests command building with random inputs
func TestBuildIRCCommand_PropertyBased(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Valid commands produce valid IRC protocol messages
	properties.Property("valid commands produce valid IRC", prop.ForAll(
		func(channel, nick, userInput string) bool {
			state := ClientState{
				CurrentChannel: channel,
				Nick:          nick,
			}
			
			commands, err := BuildIRCCommand(state, userInput)
			if err != nil {
				return false
			}
			
			// All produced commands should be valid IRC format
			for _, cmd := range commands {
				if cmd == "" {
					continue
				}
				
				// Should not contain dangerous characters
				if strings.Contains(cmd, "\r") ||
					strings.Contains(cmd, "\n") ||
					strings.Contains(cmd, "\x00") {
					return false
				}
				
				// Should have at least one word (the command)
				parts := strings.Fields(cmd)
				if len(parts) == 0 {
					return false
				}
				
				// First part should be a valid IRC command
				command := parts[0]
				if command == "" {
					return false
				}
			}
			
			return true
		},
		testutil.IRCChannelGenerator(),
		testutil.IRCNickGenerator(),
		testutil.UserInputGenerator(),
	))

	// Property: Empty input produces empty output
	properties.Property("empty input produces empty output", prop.ForAll(
		func(channel, nick string) bool {
			state := ClientState{
				CurrentChannel: channel,
				Nick:          nick,
			}
			
			commands, err := BuildIRCCommand(state, "")
			return err == nil && len(commands) == 0
		},
		testutil.IRCChannelGenerator(),
		testutil.IRCNickGenerator(),
	))

	properties.TestingRun(t)
}

// Benchmark command building performance
func BenchmarkBuildIRCCommand_PropertyBased(b *testing.B) {
	// Test data
	testCases := []struct {
		state ClientState
		input string
	}{
		{ClientState{CurrentChannel: "#test", Nick: "testnick"}, "/join #channel"},
		{ClientState{CurrentChannel: "#test", Nick: "testnick"}, "/part"},
		{ClientState{CurrentChannel: "#test", Nick: "testnick"}, "hello world"},
		{ClientState{CurrentChannel: "", Nick: "testnick"}, "/nick newnick"},
		{ClientState{CurrentChannel: "", Nick: "testnick"}, "/quit"},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		tc := testCases[i%len(testCases)]
		BuildIRCCommand(tc.state, tc.input)
	}
}