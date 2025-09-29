package testutil

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

// MessageStruct represents a simple IRC message for testing
type MessageStruct struct {
	Prefix  string
	Command string
	Params  []string
}

// IRCNickGenerator creates a gopter generator for valid IRC nicknames
func IRCNickGenerator() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		rng := rand.New(rand.NewSource(genParams.Rng.Int63()))
		nick := GenerateValidIRCNick(rng)
		
		return gopter.NewGenResult(nick, gopter.NoShrinker)
	}
}

// IRCChannelGenerator creates a gopter generator for valid IRC channel names
func IRCChannelGenerator() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		rng := rand.New(rand.NewSource(genParams.Rng.Int63()))
		channel := GenerateValidIRCChannel(rng)
		
		return gopter.NewGenResult(channel, gopter.NoShrinker)
	}
}

// IRCCommandGenerator creates a gopter generator for IRC commands
func IRCCommandGenerator() gopter.Gen {
	commands := []interface{}{
		"PRIVMSG", "JOIN", "PART", "QUIT", "NICK", "USER", "PING", "PONG", 
		"NOTICE", "MODE", "TOPIC", "KICK", "INVITE", "WHO", "WHOIS", "WHOWAS",
		"001", "002", "003", "004", "005", "401", "403", "404",
	}
	
	return gen.OneConstOf(commands...)
}

// BooleanGenerator creates a generator for boolean values
func BooleanGenerator() gopter.Gen {
	return gen.Bool()
}

// PositiveIntGenerator creates a generator for positive integers
func PositiveIntGenerator(max int) gopter.Gen {
	return gen.IntRange(1, max)
}

// PortGenerator creates a generator for valid port numbers
func PortGenerator() gopter.Gen {
	return gen.IntRange(1, 65535)
}

// HostnameGenerator creates a generator for valid hostnames
func HostnameGenerator() gopter.Gen {
	return gen.RegexMatch(`[a-z0-9][a-z0-9\-]{0,61}[a-z0-9]\.[a-z]{2,6}`)
}

// UserInputGenerator creates a generator for user input strings
func UserInputGenerator() gopter.Gen {
	inputs := []interface{}{
		"/join #test",
		"/part",
		"/quit",
		"/nick newnick",
		"/msg target hello",
		"hello world",
		"",
		"test message",
	}
	
	return gen.OneConstOf(inputs...)
}

// TimeoutGenerator creates a generator for timeout durations
func TimeoutGenerator() gopter.Gen {
	timeouts := []interface{}{
		time.Second,
		5*time.Second,
		10*time.Second,
		30*time.Second,
		time.Minute,
	}
	
	return gen.OneConstOf(timeouts...)
}

// ErrorGenerator creates a generator for various error types
func ErrorGenerator() gopter.Gen {
	errors := []interface{}{
		fmt.Errorf("connection timeout"),
		fmt.Errorf("connection refused"),
		fmt.Errorf("network unreachable"),
		fmt.Errorf("operation canceled"),
	}
	
	return gen.OneConstOf(errors...)
}

// ConnectionStateGenerator creates a generator for connection states
func ConnectionStateGenerator() gopter.Gen {
	states := []interface{}{0, 1, 2, 3, 4, 5} // StateDisconnected through StateClosed
	return gen.OneConstOf(states...)
}