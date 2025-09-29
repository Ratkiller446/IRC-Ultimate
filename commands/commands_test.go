package commands

import (
        "testing"
)

// Test the sanitize function with comprehensive edge cases
func TestSanitize(t *testing.T) {
        tests := []struct {
                name     string
                input    string
                expected string
        }{
                // Basic cases
                {
                        name:     "empty string",
                        input:    "",
                        expected: "",
                },
                {
                        name:     "normal text",
                        input:    "hello world",
                        expected: "hello world",
                },
                // Carriage return removal
                {
                        name:     "single carriage return",
                        input:    "hello\rworld",
                        expected: "helloworld",
                },
                {
                        name:     "multiple carriage returns",
                        input:    "hello\r\rworld\r",
                        expected: "helloworld",
                },
                // Newline removal
                {
                        name:     "single newline",
                        input:    "hello\nworld",
                        expected: "helloworld",
                },
                {
                        name:     "multiple newlines",
                        input:    "hello\n\nworld\n",
                        expected: "helloworld",
                },
                // Null byte removal
                {
                        name:     "single null byte",
                        input:    "hello\x00world",
                        expected: "helloworld",
                },
                {
                        name:     "multiple null bytes",
                        input:    "hello\x00\x00world\x00",
                        expected: "helloworld",
                },
                // Combined cases
                {
                        name:     "all bad characters",
                        input:    "hello\r\n\x00world\r\n\x00",
                        expected: "helloworld",
                },
                {
                        name:     "only bad characters",
                        input:    "\r\n\x00\r\n\x00",
                        expected: "",
                },
                // IRC command-like strings
                {
                        name:     "IRC privmsg with bad chars",
                        input:    "PRIVMSG #channel :Hello\rWorld\nTest\x00",
                        expected: "PRIVMSG #channel :HelloWorldTest",
                },
                // Edge cases with special characters
                {
                        name:     "unicode and special chars",
                        input:    "héllo wørld! 🎉\r\n\x00",
                        expected: "héllo wørld! 🎉",
                },
                // Long strings
                {
                        name:     "long string with scattered bad chars",
                        input:    "This is a very long string\r with some\n bad characters\x00 scattered throughout\r\n for testing purposes\x00",
                        expected: "This is a very long string with some bad characters scattered throughout for testing purposes",
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        result := Sanitize(tt.input)
                        if result != tt.expected {
                                t.Errorf("Sanitize(%q) = %q, want %q", tt.input, result, tt.expected)
                        }
                })
        }
}


// Test the buildIRCCommand function with comprehensive coverage of all IRC commands
func TestBuildIRCCommand(t *testing.T) {
        tests := []struct {
                name      string
                state     ClientState
                userInput string
                expected  []string
                expectErr bool
        }{
                // JOIN command tests
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
                        name:      "join without channel",
                        state:     ClientState{CurrentChannel: "", Nick: "testuser"},
                        userInput: "/join",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "join with malformed input",
                        state:     ClientState{CurrentChannel: "", Nick: "testuser"},
                        userInput: "/join \r\n\x00#test\r\n",
                        expected:  []string{"JOIN #test"},
                        expectErr: false,
                },

                // PART command tests
                {
                        name:      "part from current channel",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testuser"},
                        userInput: "/part",
                        expected:  []string{"PART #testchannel"},
                        expectErr: false,
                },
                {
                        name:      "part with no current channel",
                        state:     ClientState{CurrentChannel: "", Nick: "testuser"},
                        userInput: "/part",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "part with reason",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testuser"},
                        userInput: "/part Goodbye everyone!",
                        expected:  []string{"PART #testchannel :Goodbye everyone!"},
                        expectErr: false,
                },

                // NICK command tests
                {
                        name:      "nick change",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "oldnick"},
                        userInput: "/nick newnick",
                        expected:  []string{"NICK newnick"},
                        expectErr: false,
                },
                {
                        name:      "nick without parameter",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/nick",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "nick with malformed input",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/nick \r\nnewnick\x00\r",
                        expected:  []string{"NICK newnick"},
                        expectErr: false,
                },

                // MSG/PRIVMSG command tests
                {
                        name:      "msg to user",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg targetuser Hello there!",
                        expected:  []string{"PRIVMSG targetuser :Hello there!"},
                        expectErr: false,
                },
                {
                        name:      "msg to channel",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg #otherchannel Hello channel!",
                        expected:  []string{"PRIVMSG #otherchannel :Hello channel!"},
                        expectErr: false,
                },
                {
                        name:      "msg with just target",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg targetuser",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "msg without target",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "msg with malformed input",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg target\r\nHello\x00 World\r",
                        expected:  []string{"PRIVMSG targetHello :World"},
                        expectErr: false,
                },

                // QUIT command tests
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
                        name:      "quit with malformed reason",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/quit \r\nGoodbye\x00 cruel\r world\n",
                        expected:  []string{"QUIT :Goodbye cruel world"},
                        expectErr: false,
                },

                // Channel message tests (default case)
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
                        name:      "channel message with malformed input",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "Hello\r\n world\x00!",
                        expected:  []string{"PRIVMSG #testchannel :Hello world!"},
                        expectErr: false,
                },
                {
                        name:      "empty channel message",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "",
                        expected:  []string{},
                        expectErr: false,
                },

                // Unknown command tests
                {
                        name:      "unknown command",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/unknown some args",
                        expected:  []string{},
                        expectErr: false,
                },

                // Edge cases
                {
                        name:      "just slash",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "command with only spaces",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/join   ",
                        expected:  []string{},
                        expectErr: false,
                },
                {
                        name:      "mixed case commands",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/JOIN #testchannel",
                        expected:  []string{"JOIN #testchannel"},
                        expectErr: false,
                },
                {
                        name:      "command with extra spaces",
                        state:     ClientState{CurrentChannel: "#testchannel", Nick: "testnick"},
                        userInput: "/msg   target   Hello    world",
                        expected:  []string{"PRIVMSG target :Hello world"},
                        expectErr: false,
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        // Test should not panic on any input
                        var result []string
                        var err error

                        defer func() {
                                if r := recover(); r != nil {
                                        t.Errorf("BuildIRCCommand(%+v, %q) panicked: %v", tt.state, tt.userInput, r)
                                }
                        }()

                        result, err = BuildIRCCommand(tt.state, tt.userInput)

                        if tt.expectErr && err == nil {
                                t.Errorf("BuildIRCCommand(%+v, %q) expected error, got nil", tt.state, tt.userInput)
                        }
                        if !tt.expectErr && err != nil {
                                t.Errorf("BuildIRCCommand(%+v, %q) unexpected error: %v", tt.state, tt.userInput, err)
                        }

                        if len(result) != len(tt.expected) {
                                t.Errorf("BuildIRCCommand(%+v, %q) returned %d commands, want %d\nGot: %v\nWant: %v",
                                        tt.state, tt.userInput, len(result), len(tt.expected), result, tt.expected)
                                return
                        }

                        for i, cmd := range result {
                                if cmd != tt.expected[i] {
                                        t.Errorf("BuildIRCCommand(%+v, %q)[%d] = %q, want %q",
                                                tt.state, tt.userInput, i, cmd, tt.expected[i])
                                }
                        }
                })
        }
}

// Benchmark tests for performance verification
func BenchmarkSanitize(b *testing.B) {
        input := "Hello\r\nworld\x00 this is a test string with some bad characters\r\n\x00"
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
                Sanitize(input)
        }
}

func BenchmarkBuildIRCCommand(b *testing.B) {
        state := ClientState{CurrentChannel: "#testchannel", Nick: "testnick"}
        input := "/msg target Hello world this is a test message!"
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
                BuildIRCCommand(state, input)
        }
}