package parser

import (
        "strings"
        "testing"

        "github.com/leanovate/gopter"
        "github.com/leanovate/gopter/prop"
        "irc-client/testutil"
)

// TestParseMessage_PropertyBased uses property-based testing to verify IRC message parsing
func TestParseMessage_PropertyBased(t *testing.T) {
        properties := gopter.NewProperties(nil)

        // Property: Valid IRC messages always have non-empty commands
        properties.Property("valid IRC messages have non-empty commands", prop.ForAll(
                func(command string) bool {
                        if command == "" {
                                return true // Skip empty commands
                        }
                        
                        parsed := ParseMessage(command)
                        return parsed.Command != ""
                },
                testutil.IRCCommandGenerator(),
        ))

        // Property: Parameters never contain CR, LF, or NULL bytes
        properties.Property("parsed messages are clean", prop.ForAll(
                func(input string) bool {
                        parsed := ParseMessage(input)
                        
                        // Check prefix
                        if strings.Contains(parsed.Prefix, "\r") ||
                                strings.Contains(parsed.Prefix, "\n") ||
                                strings.Contains(parsed.Prefix, "\x00") {
                                return false
                        }
                        
                        // Check command
                        if strings.Contains(parsed.Command, "\r") ||
                                strings.Contains(parsed.Command, "\n") ||
                                strings.Contains(parsed.Command, "\x00") {
                                return false
                        }
                        
                        // Check parameters
                        for _, param := range parsed.Params {
                                if strings.Contains(param, "\r") ||
                                        strings.Contains(param, "\n") ||
                                        strings.Contains(param, "\x00") {
                                        return false
                                }
                        }
                        return true
                },
                testutil.UserInputGenerator(),
        ))

        // Property: Empty string parsing
        properties.Property("empty strings are handled gracefully", prop.ForAll(
                func() bool {
                        parsed := ParseMessage("")
                        return parsed.Command == "" && len(parsed.Params) == 0
                },
        ))

        // Property: Command format validation for known commands
        properties.Property("command format is valid for known commands", prop.ForAll(
                func(command string) bool {
                        parsed := ParseMessage(command)
                        
                        if parsed.Command == "" {
                                return true
                        }
                        
                        // Command should be alphanumeric (letters and/or numbers)
                        for _, char := range parsed.Command {
                                if !((char >= 'A' && char <= 'Z') || 
                                        (char >= 'a' && char <= 'z') || 
                                        (char >= '0' && char <= '9')) {
                                        return false
                                }
                        }
                        return true
                },
                testutil.IRCCommandGenerator(),
        ))

        // Property: Parameter count limits (RFC 1459 allows max 15 parameters)
        properties.Property("parameter count within limits", prop.ForAll(
                func(input string) bool {
                        parsed := ParseMessage(input)
                        return len(parsed.Params) <= 15
                },
                testutil.UserInputGenerator(),
        ))

        // Run all properties
        properties.TestingRun(t)
}

// TestParseMessage_RFC1459Compliance tests specific RFC 1459 compliance requirements
func TestParseMessage_RFC1459Compliance(t *testing.T) {
        properties := gopter.NewProperties(nil)

        // RFC 1459: Commands are case-insensitive (we store as-is)
        properties.Property("command case handling", prop.ForAll(
                func(command string) bool {
                        if command == "" {
                                return true
                        }
                        
                        upperCase := ParseMessage(strings.ToUpper(command))
                        lowerCase := ParseMessage(strings.ToLower(command))
                        
                        // Commands should be preserved as-is (parser doesn't change case)
                        return upperCase.Command == strings.ToUpper(command) &&
                                lowerCase.Command == strings.ToLower(command)
                },
                testutil.IRCCommandGenerator(),
        ))

        properties.TestingRun(t)
}

// Benchmark property-based testing performance
func BenchmarkParseMessage_PropertyBased(b *testing.B) {
        // Generate test data
        messages := []string{
                "PRIVMSG #channel :Hello",
                ":nick!user@host JOIN #channel",
                "PING :server.example.com",
                ":server 001 nick :Welcome",
                "QUIT :Goodbye",
        }
        
        b.ResetTimer()
        
        for i := 0; i < b.N; i++ {
                msg := messages[i%len(messages)]
                ParseMessage(msg)
        }
}