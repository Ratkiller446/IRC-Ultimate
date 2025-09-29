package testutil

import (
        "fmt"
        "math/rand"
        "strings"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

// Message represents an IRC message structure for testing
// This is a duplicate of parser.Message to avoid import cycles
type Message struct {
        Prefix  string
        Command string
        Params  []string
}

// IRCTestHelper provides utilities for IRC-specific testing
type IRCTestHelper struct {
        t      *testing.T
        assert *assert.Assertions
        require *require.Assertions
}

// NewIRCTestHelper creates a new IRC test helper
func NewIRCTestHelper(t *testing.T) *IRCTestHelper {
        return &IRCTestHelper{
                t:       t,
                assert:  assert.New(t),
                require: require.New(t),
        }
}

// AssertValidIRCMessage verifies that a message follows IRC protocol rules
func (h *IRCTestHelper) AssertValidIRCMessage(msg Message) {
        // Command should not be empty
        h.assert.NotEmpty(msg.Command, "IRC command should not be empty")
        
        // Command should be alphanumeric or numeric (RFC 1459)
        h.assert.Regexp(`^[A-Z0-9]+$`, msg.Command, "IRC command should be alphanumeric")
        
        // Prefix should not contain spaces or invalid characters
        if msg.Prefix != "" {
                h.assert.NotContains(msg.Prefix, " ", "IRC prefix should not contain spaces")
                h.assert.NotContains(msg.Prefix, "\r", "IRC prefix should not contain CR")
                h.assert.NotContains(msg.Prefix, "\n", "IRC prefix should not contain LF")
        }
        
        // Parameters should not contain invalid characters (except trailing parameter)
        for i, param := range msg.Params {
                h.assert.NotContains(param, "\r", "IRC parameter should not contain CR")
                h.assert.NotContains(param, "\n", "IRC parameter should not contain LF")
                h.assert.NotContains(param, "\x00", "IRC parameter should not contain NULL")
                
                // Only the last parameter can contain spaces (if it's a trailing parameter)
                if i < len(msg.Params)-1 {
                        h.assert.NotContains(param, " ", "IRC middle parameter should not contain spaces")
                }
        }
}

// AssertIRCMessageEquals compares two IRC messages
func (h *IRCTestHelper) AssertIRCMessageEquals(expected, actual Message) {
        h.assert.Equal(expected.Prefix, actual.Prefix, "IRC message prefix mismatch")
        h.assert.Equal(expected.Command, actual.Command, "IRC message command mismatch")
        h.assert.Equal(expected.Params, actual.Params, "IRC message parameters mismatch")
}

// GenerateValidIRCNick generates a valid IRC nickname
func GenerateValidIRCNick(rng *rand.Rand) string {
        // IRC nicknames: 1-30 chars, start with letter, contain letters/numbers/special chars
        validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789[]{}\\|`_-^"
        firstChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ[]{}\\|`_-^"
        
        length := rng.Intn(30) + 1 // 1-30 characters
        nick := string(firstChars[rng.Intn(len(firstChars))])
        
        for i := 1; i < length; i++ {
                nick += string(validChars[rng.Intn(len(validChars))])
        }
        
        return nick
}

// GenerateValidIRCChannel generates a valid IRC channel name
func GenerateValidIRCChannel(rng *rand.Rand) string {
        // IRC channels start with # or & and can contain most characters except space, comma, ^G
        prefixes := []string{"#", "&"}
        validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789[]{}\\|`_-=+;:<>?/."
        
        prefix := prefixes[rng.Intn(len(prefixes))]
        length := rng.Intn(49) + 1 // 1-49 characters after prefix (50 char limit)
        
        channel := prefix
        for i := 0; i < length; i++ {
                channel += string(validChars[rng.Intn(len(validChars))])
        }
        
        return channel
}

// GenerateRandomString generates a random string with given character set
func GenerateRandomString(rng *rand.Rand, charset string, minLen, maxLen int) string {
        length := rng.Intn(maxLen-minLen+1) + minLen
        result := make([]byte, length)
        for i := range result {
                result[i] = charset[rng.Intn(len(charset))]
        }
        return string(result)
}

// GenerateIRCMessage generates a random IRC message for property-based testing
func GenerateIRCMessage(rng *rand.Rand) Message {
        msg := Message{
                Params: []string{},
        }
        
        // Random prefix (30% chance)
        if rng.Float32() < 0.3 {
                if rng.Float32() < 0.5 {
                        // Simple nick prefix
                        msg.Prefix = GenerateValidIRCNick(rng)
                } else {
                        // Full user prefix (nick!user@host)
                        nick := GenerateValidIRCNick(rng)
                        user := GenerateRandomString(rng, "abcdefghijklmnopqrstuvwxyz0123456789", 1, 10)
                        host := fmt.Sprintf("%s.%s", 
                                GenerateRandomString(rng, "abcdefghijklmnopqrstuvwxyz", 3, 8),
                                GenerateRandomString(rng, "abcdefghijklmnopqrstuvwxyz", 2, 4))
                        msg.Prefix = fmt.Sprintf("%s!%s@%s", nick, user, host)
                }
        }
        
        // Random command
        commands := []string{"PRIVMSG", "JOIN", "PART", "QUIT", "NICK", "USER", "PING", "PONG", "NOTICE", "MODE"}
        numerics := []string{"001", "002", "003", "004", "005", "250", "251", "252", "253", "254", "255", "401", "403", "404"}
        
        if rng.Float32() < 0.7 {
                msg.Command = commands[rng.Intn(len(commands))]
        } else {
                msg.Command = numerics[rng.Intn(len(numerics))]
        }
        
        // Random parameters (0-15 parameters as per RFC)
        paramCount := rng.Intn(16)
        for i := 0; i < paramCount; i++ {
                if i == paramCount-1 && rng.Float32() < 0.4 {
                        // Last parameter can be a trailing parameter with spaces
                        param := GenerateRandomString(rng, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !@#$%^&*()_+-=[]{}|;':\",./<>?", 0, 100)
                        msg.Params = append(msg.Params, param)
                } else {
                        // Middle parameter without spaces
                        param := GenerateRandomString(rng, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;':\",./<>?", 1, 50)
                        msg.Params = append(msg.Params, param)
                }
        }
        
        return msg
}

// BuildIRCMessageString builds a raw IRC message string from a Message struct
func BuildIRCMessageString(msg Message) string {
        var parts []string
        
        // Add prefix if present
        if msg.Prefix != "" {
                parts = append(parts, ":"+msg.Prefix)
        }
        
        // Add command
        parts = append(parts, msg.Command)
        
        // Add parameters
        for i, param := range msg.Params {
                if i == len(msg.Params)-1 && (strings.Contains(param, " ") || param == "" || strings.HasPrefix(param, ":")) {
                        // Last parameter with spaces or empty or starting with colon needs trailing format
                        parts = append(parts, ":"+param)
                } else {
                        parts = append(parts, param)
                }
        }
        
        return strings.Join(parts, " ")
}

// GenerateInvalidIRCMessage generates an intentionally malformed IRC message
func GenerateInvalidIRCMessage(rng *rand.Rand) string {
        invalidPatterns := []func() string{
                // Empty message
                func() string { return "" },
                // Only spaces
                func() string { return "   " },
                // Prefix without command
                func() string { return ":nick" },
                // Multiple prefixes
                func() string { return ":nick1 :nick2 PRIVMSG #channel :hello" },
                // Invalid characters in command
                func() string { return "PRIV MSG #channel :hello" },
                // Command with numbers only (should be valid actually, so this is a false invalid)
                func() string { return "123 param" },
                // Very long line (over 512 bytes)
                func() string { 
                        return "PRIVMSG #channel :" + strings.Repeat("x", 600)
                },
                // Embedded null bytes
                func() string { return "PRIVMSG #channel :hello\x00world" },
                // Embedded CR/LF
                func() string { return "PRIVMSG #channel :hello\r\nworld" },
        }
        
        pattern := invalidPatterns[rng.Intn(len(invalidPatterns))]
        return pattern()
}

// CommonIRCScenarios provides common IRC message scenarios for testing
var CommonIRCScenarios = []struct {
        Name    string
        Message string
        Parsed  Message
}{
        {
                Name:    "Simple QUIT",
                Message: "QUIT",
                Parsed: Message{
                        Prefix:  "",
                        Command: "QUIT",
                        Params:  []string{},
                },
        },
        {
                Name:    "PRIVMSG to channel",
                Message: ":nick!user@host PRIVMSG #channel :Hello world!",
                Parsed: Message{
                        Prefix:  "nick!user@host",
                        Command: "PRIVMSG",
                        Params:  []string{"#channel", "Hello world!"},
                },
        },
        {
                Name:    "Numeric reply",
                Message: ":server.example.com 001 nick :Welcome to the network",
                Parsed: Message{
                        Prefix:  "server.example.com",
                        Command: "001",
                        Params:  []string{"nick", "Welcome to the network"},
                },
        },
        {
                Name:    "JOIN with key",
                Message: "JOIN #channel secretkey",
                Parsed: Message{
                        Prefix:  "",
                        Command: "JOIN",
                        Params:  []string{"#channel", "secretkey"},
                },
        },
        {
                Name:    "PING message",
                Message: "PING :server.example.com",
                Parsed: Message{
                        Prefix:  "",
                        Command: "PING",
                        Params:  []string{"server.example.com"},
                },
        },
}

// WaitWithTimeout waits for a condition with timeout
func WaitWithTimeout(t *testing.T, condition func() bool, timeout time.Duration, message string) {
        deadline := time.Now().Add(timeout)
        for time.Now().Before(deadline) {
                if condition() {
                        return
                }
                time.Sleep(10 * time.Millisecond)
        }
        t.Fatalf("Timeout waiting for condition: %s", message)
}

// AssertEventuallyTrue asserts that a condition becomes true within a timeout
func AssertEventuallyTrue(t *testing.T, condition func() bool, timeout time.Duration, message string) {
        WaitWithTimeout(t, condition, timeout, message)
}