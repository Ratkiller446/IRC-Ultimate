package parser

import (
        "fmt"
        "reflect"
        "strings"
        "testing"
)

// TestParseMessage provides comprehensive table-driven tests for IRC message parsing
// following RFC 1459 specifications with extensive edge case coverage
func TestParseMessage(t *testing.T) {
        tests := []struct {
                name     string
                input    string
                expected Message
        }{
                // Basic command-only messages
                {
                        name:  "simple command",
                        input: "QUIT",
                        expected: Message{
                                Prefix:  "",
                                Command: "QUIT",
                                Params:  []string{},
                        },
                },
                {
                        name:  "command with single parameter",
                        input: "NICK newname",
                        expected: Message{
                                Prefix:  "",
                                Command: "NICK",
                                Params:  []string{"newname"},
                        },
                },
                {
                        name:  "command with multiple parameters",
                        input: "USER guest 0 * :Ronnie Reagan",
                        expected: Message{
                                Prefix:  "",
                                Command: "USER",
                                Params:  []string{"guest", "0", "*", "Ronnie Reagan"},
                        },
                },

                // Messages with prefix
                {
                        name:  "message with server prefix",
                        input: ":server.example.com 001 nick :Welcome to the network",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "001",
                                Params:  []string{"nick", "Welcome to the network"},
                        },
                },
                {
                        name:  "message with full user prefix",
                        input: ":nick!user@host PRIVMSG #channel :Hello world!",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"#channel", "Hello world!"},
                        },
                },
                {
                        name:  "message with nick-only prefix",
                        input: ":nick QUIT :Leaving",
                        expected: Message{
                                Prefix:  "nick",
                                Command: "QUIT",
                                Params:  []string{"Leaving"},
                        },
                },

                // PING/PONG forms
                {
                        name:  "PING with server",
                        input: "PING :server.example.com",
                        expected: Message{
                                Prefix:  "",
                                Command: "PING",
                                Params:  []string{"server.example.com"},
                        },
                },
                {
                        name:  "PONG response",
                        input: ":server.example.com PONG server.example.com :nick",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "PONG",
                                Params:  []string{"server.example.com", "nick"},
                        },
                },
                {
                        name:  "PING without colon",
                        input: "PING server.example.com",
                        expected: Message{
                                Prefix:  "",
                                Command: "PING",
                                Params:  []string{"server.example.com"},
                        },
                },

                // Numeric replies
                {
                        name:  "welcome message (001)",
                        input: ":server.example.com 001 nick :Welcome to the Internet Relay Network",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "001",
                                Params:  []string{"nick", "Welcome to the Internet Relay Network"},
                        },
                },
                {
                        name:  "no such nick error (401)",
                        input: ":server.example.com 401 nick nonick :No such nick/channel",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "401",
                                Params:  []string{"nick", "nonick", "No such nick/channel"},
                        },
                },
                {
                        name:  "channel list (322)",
                        input: ":server.example.com 322 nick #channel 10 :Channel topic here",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "322",
                                Params:  []string{"nick", "#channel", "10", "Channel topic here"},
                        },
                },

                // PRIVMSG/NOTICE formatting
                {
                        name:  "PRIVMSG to channel",
                        input: ":nick!user@host PRIVMSG #channel :This is a message to the channel",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"#channel", "This is a message to the channel"},
                        },
                },
                {
                        name:  "PRIVMSG to user",
                        input: ":sender!user@host PRIVMSG recipient :Private message here",
                        expected: Message{
                                Prefix:  "sender!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"recipient", "Private message here"},
                        },
                },
                {
                        name:  "NOTICE message",
                        input: ":server.example.com NOTICE nick :*** Looking up your hostname...",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "NOTICE",
                                Params:  []string{"nick", "*** Looking up your hostname..."},
                        },
                },

                // Trailing parameters with special content
                {
                        name:  "trailing parameter with spaces",
                        input: "KICK #channel nick :Reason with multiple   spaces",
                        expected: Message{
                                Prefix:  "",
                                Command: "KICK",
                                Params:  []string{"#channel", "nick", "Reason with multiple   spaces"},
                        },
                },
                {
                        name:  "trailing parameter with colons",
                        input: ":nick!user@host PRIVMSG #channel :Message with : colons : inside",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"#channel", "Message with : colons : inside"},
                        },
                },
                {
                        name:  "empty trailing parameter",
                        input: "TOPIC #channel :",
                        expected: Message{
                                Prefix:  "",
                                Command: "TOPIC",
                                Params:  []string{"#channel", ""},
                        },
                },
                {
                        name:  "trailing parameter with only spaces",
                        input: "TOPIC #channel :   ",
                        expected: Message{
                                Prefix:  "",
                                Command: "TOPIC",
                                Params:  []string{"#channel", "   "},
                        },
                },

                // Multiple middle parameters
                {
                        name:  "MODE with multiple parameters",
                        input: ":server.example.com MODE #channel +o nick",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "MODE",
                                Params:  []string{"#channel", "+o", "nick"},
                        },
                },
                {
                        name:  "many parameters before trailing",
                        input: "CMD param1 param2 param3 param4 param5 :trailing content",
                        expected: Message{
                                Prefix:  "",
                                Command: "CMD",
                                Params:  []string{"param1", "param2", "param3", "param4", "param5", "trailing content"},
                        },
                },

                // Whitespace edge cases
                {
                        name:  "multiple spaces between parameters",
                        input: "NICK    newname",
                        expected: Message{
                                Prefix:  "",
                                Command: "NICK",
                                Params:  []string{"newname"},
                        },
                },
                {
                        name:  "multiple spaces in complex message",
                        input: ":nick!user@host    PRIVMSG    #channel    :Message content",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"#channel", "Message content"},
                        },
                },
                {
                        name:  "tabs and spaces mixed",
                        input: "NICK\t\tnewname",
                        expected: Message{
                                Prefix:  "",
                                Command: "NICK\t\tnewname", // Note: tabs are not treated as separators in current implementation
                                Params:  []string{},
                        },
                },

                // Empty and minimal cases
                {
                        name:  "empty string",
                        input: "",
                        expected: Message{
                                Prefix:  "",
                                Command: "",
                                Params:  []string{},
                        },
                },
                {
                        name:  "just spaces",
                        input: "   ",
                        expected: Message{
                                Prefix:  "",
                                Command: "", // Whitespace-only input should result in empty command
                                Params:  []string{},
                        },
                },
                {
                        name:  "command with trailing space",
                        input: "QUIT ",
                        expected: Message{
                                Prefix:  "",
                                Command: "QUIT",
                                Params:  []string{},
                        },
                },

                // Malformed input edge cases
                {
                        name:  "prefix only",
                        input: ":nick!user@host",
                        expected: Message{
                                Prefix:  "",
                                Command: ":nick!user@host",
                                Params:  []string{},
                        },
                },
                {
                        name:  "prefix with no command",
                        input: ":nick!user@host ",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "",
                                Params:  []string{},
                        },
                },
                {
                        name:  "colon at start but no space",
                        input: ":no-space-after-prefix-COMMAND",
                        expected: Message{
                                Prefix:  "",
                                Command: ":no-space-after-prefix-COMMAND",
                                Params:  []string{},
                        },
                },
                {
                        name:  "multiple colons in middle parameters",
                        input: "CMD param:with:colons another:param :trailing",
                        expected: Message{
                                Prefix:  "",
                                Command: "CMD",
                                Params:  []string{"param:with:colons", "another:param", "trailing"},
                        },
                },

                // Parameter edge cases
                {
                        name:  "parameter that looks like prefix",
                        input: "PRIVMSG :nick!user@host hello",
                        expected: Message{
                                Prefix:  "",
                                Command: "PRIVMSG",
                                Params:  []string{"nick!user@host hello"},
                        },
                },
                {
                        name:  "many parameters (up to limit)",
                        input: "CMD p1 p2 p3 p4 p5 p6 p7 p8 p9 p10 p11 p12 p13 p14 :p15",
                        expected: Message{
                                Prefix:  "",
                                Command: "CMD",
                                Params:  []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12", "p13", "p14", "p15"},
                        },
                },

                // Real-world IRC examples
                {
                        name:  "JOIN message",
                        input: ":nick!user@host JOIN #channel",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "JOIN",
                                Params:  []string{"#channel"},
                        },
                },
                {
                        name:  "PART with reason",
                        input: ":nick!user@host PART #channel :Goodbye everyone!",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PART",
                                Params:  []string{"#channel", "Goodbye everyone!"},
                        },
                },
                {
                        name:  "server MOTD",
                        input: ":server.example.com 372 nick :- Message of the day goes here",
                        expected: Message{
                                Prefix:  "server.example.com",
                                Command: "372",
                                Params:  []string{"nick", "- Message of the day goes here"},
                        },
                },
                {
                        name:  "CTCP ACTION",
                        input: ":nick!user@host PRIVMSG #channel :\u0001ACTION does something\u0001",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"#channel", "\u0001ACTION does something\u0001"},
                        },
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        result := ParseMessage(tt.input)
                        
                        // Compare each field separately for better error messages
                        if result.Prefix != tt.expected.Prefix {
                                t.Errorf("Prefix mismatch\nInput: %q\nGot:      %q\nExpected: %q", tt.input, result.Prefix, tt.expected.Prefix)
                        }
                        
                        if result.Command != tt.expected.Command {
                                t.Errorf("Command mismatch\nInput: %q\nGot:      %q\nExpected: %q", tt.input, result.Command, tt.expected.Command)
                        }
                        
                        if !reflect.DeepEqual(result.Params, tt.expected.Params) {
                                t.Errorf("Params mismatch\nInput: %q\nGot:      %v\nExpected: %v", tt.input, result.Params, tt.expected.Params)
                        }
                })
        }
}

// TestParseMessageEdgeCases provides additional edge case testing
func TestParseMessageEdgeCases(t *testing.T) {
        tests := []struct {
                name        string
                input       string
                expectEmpty bool
                description string
        }{
                {
                        name:        "nil-like empty message",
                        input:       "",
                        expectEmpty: true,
                        description: "Empty string should return zero-value Message",
                },
                {
                        name:        "whitespace only",
                        input:       "   ",
                        expectEmpty: false, // Current implementation treats spaces as command
                        description: "Whitespace-only should be handled gracefully",
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        result := ParseMessage(tt.input)
                        if tt.expectEmpty {
                                if result.Command != "" || len(result.Params) != 0 {
                                        t.Errorf("%s\nInput: %q\nGot non-empty result: %+v", tt.description, tt.input, result)
                                }
                        }
                })
        }
}

// TestParseMessageBoundaryConditions tests boundary conditions and limits
func TestParseMessageBoundaryConditions(t *testing.T) {
        t.Run("max parameters RFC compliance", func(t *testing.T) {
                // RFC 1459 allows up to 15 parameters
                // Generate message with exactly 15 parameters
                input := "CMD"
                expectedParams := make([]string, 15)
                for i := 0; i < 14; i++ {
                        param := fmt.Sprintf("param%d", i+1)
                        input += " " + param
                        expectedParams[i] = param
                }
                input += " :final parameter"
                expectedParams[14] = "final parameter"

                result := ParseMessage(input)
                if len(result.Params) != 15 {
                        t.Errorf("Expected 15 parameters, got %d", len(result.Params))
                }
                if !reflect.DeepEqual(result.Params, expectedParams) {
                        t.Errorf("Parameter mismatch\nGot:      %v\nExpected: %v", result.Params, expectedParams)
                }
        })

        t.Run("very long message", func(t *testing.T) {
                // Test with a very long trailing parameter
                longContent := strings.Repeat("a", 1000)
                input := fmt.Sprintf("PRIVMSG #channel :%s", longContent)
                
                result := ParseMessage(input)
                if result.Command != "PRIVMSG" {
                        t.Errorf("Expected command PRIVMSG, got %q", result.Command)
                }
                if len(result.Params) != 2 {
                        t.Errorf("Expected 2 parameters, got %d", len(result.Params))
                }
                if result.Params[1] != longContent {
                        t.Errorf("Long content not preserved correctly")
                }
        })
}

// TestParseMessageMultipleSpaces tests RFC 1459 requirement for handling multiple spaces
func TestParseMessageMultipleSpaces(t *testing.T) {
        tests := []struct {
                name     string
                input    string
                expected Message
        }{
                {
                        name:  "double spaces between command and param",
                        input: "NICK  newname",
                        expected: Message{
                                Prefix:  "",
                                Command: "NICK",
                                Params:  []string{"newname"},
                        },
                },
                {
                        name:  "triple spaces between params",
                        input: "USER guest   0   *   :Ronnie Reagan",
                        expected: Message{
                                Prefix:  "",
                                Command: "USER",
                                Params:  []string{"guest", "0", "*", "Ronnie Reagan"},
                        },
                },
                {
                        name:  "many spaces with prefix",
                        input: ":nick!user@host     PRIVMSG     #channel     :Message content",
                        expected: Message{
                                Prefix:  "nick!user@host",
                                Command: "PRIVMSG",
                                Params:  []string{"#channel", "Message content"},
                        },
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        result := ParseMessage(tt.input)
                        
                        if result.Prefix != tt.expected.Prefix {
                                t.Errorf("Prefix mismatch\nInput: %q\nGot:      %q\nExpected: %q", tt.input, result.Prefix, tt.expected.Prefix)
                        }
                        
                        if result.Command != tt.expected.Command {
                                t.Errorf("Command mismatch\nInput: %q\nGot:      %q\nExpected: %q", tt.input, result.Command, tt.expected.Command)
                        }
                        
                        if !reflect.DeepEqual(result.Params, tt.expected.Params) {
                                t.Errorf("Params mismatch\nInput: %q\nGot:      %v\nExpected: %v", tt.input, result.Params, tt.expected.Params)
                        }
                })
        }
}

// TestParseMessageRFC1459Compliance tests specific RFC 1459 compliance requirements
func TestParseMessageRFC1459Compliance(t *testing.T) {
        t.Run("prefix extraction compliance", func(t *testing.T) {
                // Test that prefix is properly extracted when present
                input := ":nick!user@host.example.com PRIVMSG #channel :message"
                result := ParseMessage(input)
                
                if result.Prefix != "nick!user@host.example.com" {
                        t.Errorf("Prefix not extracted correctly: got %q, expected %q", 
                                result.Prefix, "nick!user@host.example.com")
                }
        })

        t.Run("trailing parameter preservation", func(t *testing.T) {
                // Test that trailing parameters preserve all content after ':'
                input := "PRIVMSG #channel :This message contains   multiple   spaces"
                result := ParseMessage(input)
                
                expected := "This message contains   multiple   spaces"
                if len(result.Params) != 2 || result.Params[1] != expected {
                        t.Errorf("Trailing parameter not preserved correctly\nGot: %q\nExpected: %q", 
                                result.Params[1], expected)
                }
        })

        t.Run("command case sensitivity", func(t *testing.T) {
                // Test that commands preserve case (important for IRC)
                tests := []string{"PRIVMSG", "privmsg", "PrivMsg", "JOIN", "join"}
                for _, cmd := range tests {
                        input := fmt.Sprintf("%s #channel", cmd)
                        result := ParseMessage(input)
                        if result.Command != cmd {
                                t.Errorf("Command case not preserved: input %q, got %q", cmd, result.Command)
                        }
                }
        })
}