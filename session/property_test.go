package session

import (
        "testing"

        "github.com/leanovate/gopter"
        "github.com/leanovate/gopter/gen"
        "github.com/leanovate/gopter/prop"

        "irc-client/parser"
)

// Property test: Nick changes should always be consistent
func TestPropertyNickConsistency(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("nick changes are consistent", prop.ForAll(
                func(initialNick, newNick string) bool {
                        manager := NewManager(initialNick, "irc.example.com", 6667)
                        
                        // Initial nick should match
                        if manager.GetCurrentNick() != initialNick {
                                return false
                        }
                        
                        // Change nick
                        manager.SetCurrentNick(newNick)
                        
                        // New nick should match
                        if manager.GetCurrentNick() != newNick {
                                return false
                        }
                        
                        // Client state should reflect the change
                        state := manager.GetClientState()
                        return state.Nick == newNick
                },
                gen.AlphaString(),
                gen.AlphaString(),
        ))

        properties.TestingRun(t)
}

// Property test: Channel state should be consistent
func TestPropertyChannelConsistency(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("channel state is consistent", prop.ForAll(
                func(channels []string) bool {
                        manager := NewManager("testnick", "irc.example.com", 6667)
                        
                        // Should start with no channel
                        if manager.GetCurrentChannel() != "" {
                                return false
                        }
                        
                        var lastChannel string
                        for _, channel := range channels {
                                // Join channel
                                manager.JoinChannel(channel)
                                
                                // Current channel should match
                                if manager.GetCurrentChannel() != channel {
                                        return false
                                }
                                
                                // Client state should reflect the change
                                state := manager.GetClientState()
                                if state.CurrentChannel != channel {
                                        return false
                                }
                                
                                lastChannel = channel
                        }
                        
                        // Part from last channel
                        if lastChannel != "" {
                                partedChannel := manager.PartChannel()
                                
                                // Should return the channel we parted from
                                if partedChannel != lastChannel {
                                        return false
                                }
                                
                                // Should have no current channel
                                if manager.GetCurrentChannel() != "" {
                                        return false
                                }
                                
                                // Client state should reflect the change
                                state := manager.GetClientState()
                                if state.CurrentChannel != "" {
                                        return false
                                }
                        }
                        
                        return true
                },
                gen.SliceOf(gen.OneGenOf(
                        gen.Const("#test"),
                        gen.Const("#random"),
                        gen.Const("#channel"),
                )),
        ))

        properties.TestingRun(t)
}

// Property test: HandleMessage should never panic
func TestPropertyHandleMessageNeverPanics(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("HandleMessage never panics", prop.ForAll(
                func(command string, prefix string, params []string) bool {
                        manager := NewManager("testnick", "irc.example.com", 6667)
                        
                        msg := &parser.Message{
                                Command: command,
                                Prefix:  prefix,
                                Params:  params,
                        }
                        
                        // This should never panic, regardless of input
                        defer func() {
                                if r := recover(); r != nil {
                                        t.Errorf("HandleMessage panicked with input: command=%s, prefix=%s, params=%v", command, prefix, params)
                                }
                        }()
                        
                        manager.HandleMessage(msg)
                        return true
                },
                gen.OneGenOf(gen.Const("JOIN"), gen.Const("PART"), gen.Const("NICK"), gen.Const("PRIVMSG"), gen.Const("UNKNOWN")),
                gen.AlphaString(),
                gen.SliceOf(gen.AlphaString()),
        ))

        properties.TestingRun(t)
}

// Property test: JOIN/PART message handling
func TestPropertyJoinPartHandling(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("JOIN/PART messages handled correctly", prop.ForAll(
                func(nick, channel string) bool {
                        manager := NewManager(nick, "irc.example.com", 6667)
                        
                        // Test JOIN message for self
                        joinMsg := &parser.Message{
                                Command: "JOIN",
                                Prefix:  nick + "!user@host.com",
                                Params:  []string{channel},
                        }
                        
                        err := manager.HandleMessage(joinMsg)
                        if err != nil {
                                return false
                        }
                        
                        // Should be in the channel now
                        if manager.GetCurrentChannel() != channel {
                                return false
                        }
                        
                        // Test PART message for self
                        partMsg := &parser.Message{
                                Command: "PART",
                                Prefix:  nick + "!user@host.com",
                                Params:  []string{channel},
                        }
                        
                        err = manager.HandleMessage(partMsg)
                        if err != nil {
                                return false
                        }
                        
                        // Should not be in any channel now
                        return manager.GetCurrentChannel() == ""
                },
                gen.AlphaString(),
                gen.Const("#testchan"),
        ))

        properties.TestingRun(t)
}

// Property test: NICK message handling
func TestPropertyNickHandling(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("NICK messages handled correctly", prop.ForAll(
                func(oldNick, newNick string) bool {
                        manager := NewManager(oldNick, "irc.example.com", 6667)
                        
                        // Test NICK message for self
                        nickMsg := &parser.Message{
                                Command: "NICK",
                                Prefix:  oldNick + "!user@host.com",
                                Params:  []string{newNick},
                        }
                        
                        err := manager.HandleMessage(nickMsg)
                        if err != nil {
                                return false
                        }
                        
                        // Should have new nick now
                        return manager.GetCurrentNick() == newNick
                },
                gen.AlphaString(),
                gen.AlphaString(),
        ))

        properties.TestingRun(t)
}

// Property test: Connection state consistency
func TestPropertyConnectionState(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("connection state is consistent", prop.ForAll(
                func(states []bool) bool {
                        manager := NewManager("testnick", "irc.example.com", 6667)
                        
                        // Should start disconnected
                        if manager.IsConnected() {
                                return false
                        }
                        
                        for _, state := range states {
                                manager.SetConnected(state)
                                
                                // Connection state should match
                                if manager.IsConnected() != state {
                                        return false
                                }
                                
                                // Client state should reflect the change
                                clientState := manager.GetClientState()
                                if clientState.Connected != state {
                                        return false
                                }
                        }
                        
                        return true
                },
                gen.SliceOf(gen.Bool()),
        ))

        properties.TestingRun(t)
}

// Property test: extractNickFromPrefix function
func TestPropertyExtractNickFromPrefix(t *testing.T) {
        properties := gopter.NewProperties(nil)

        properties.Property("extractNickFromPrefix extracts correctly", prop.ForAll(
                func(nick, user, host string) bool {
                        // Test full prefix format
                        fullPrefix := nick + "!" + user + "@" + host
                        extracted := extractNickFromPrefix(fullPrefix)
                        if extracted != nick {
                                return false
                        }
                        
                        // Test nick-only format
                        extracted = extractNickFromPrefix(nick)
                        if extracted != nick {
                                return false
                        }
                        
                        // Test empty prefix
                        extracted = extractNickFromPrefix("")
                        return extracted == ""
                },
                gen.AlphaString(),
                gen.AlphaString(),
                gen.AlphaString(),
        ))

        properties.TestingRun(t)
}