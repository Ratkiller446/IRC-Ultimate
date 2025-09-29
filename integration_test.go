//go:build integration
// +build integration

package main

import (
        "bufio"
        "context"
        "crypto/tls"
        "fmt"
        "net"
        "os"
        "strings"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        "irc-client/conn"
        "irc-client/parser"
)

// Integration tests require a real IRC server running
// Set IRC_TEST_SERVER environment variable to specify server
// Example: IRC_TEST_SERVER=irc.libera.chat:6667

func getTestServer() string {
        server := os.Getenv("IRC_TEST_SERVER")
        if server == "" {
                return "localhost:6667"
        }
        return server
}

func getTLSTestServer() string {
        server := os.Getenv("IRC_TEST_TLS_SERVER")
        if server == "" {
                return "localhost:6697"
        }
        return server
}

func TestIntegration_BasicConnection(t *testing.T) {
        server := getTestServer()
        
        config := conn.Config{
                Server:  strings.Split(server, ":")[0],
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        // Test basic connection
        connection, err := conn.Connect(config)
        if err != nil {
                t.Skipf("Skipping integration test - cannot connect to %s: %v", server, err)
        }
        defer connection.Close()
        
        // Verify connection is established
        assert.NotNil(t, connection)
        
        // Test writing to connection
        _, err = connection.Write([]byte("PING :test\r\n"))
        require.NoError(t, err)
        
        // Test reading from connection
        connection.SetReadDeadline(time.Now().Add(5 * time.Second))
        buffer := make([]byte, 1024)
        n, err := connection.Read(buffer)
        
        if err == nil {
                response := string(buffer[:n])
                t.Logf("Server response: %s", response)
        }
}

func TestIntegration_TLSConnection(t *testing.T) {
        server := getTLSTestServer()
        
        config := conn.Config{
                Server:   strings.Split(server, ":")[0],
                Port:     6697,
                TLS:      true,
                Timeout:  10 * time.Second,
                Insecure: true, // For test servers with self-signed certs
        }
        
        // Test TLS connection
        connection, err := conn.Connect(config)
        if err != nil {
                t.Skipf("Skipping TLS integration test - cannot connect to %s: %v", server, err)
        }
        defer connection.Close()
        
        // Verify connection is established
        assert.NotNil(t, connection)
        
        // Test that it's actually a TLS connection
        _, isTLS := connection.(*tls.Conn)
        assert.True(t, isTLS, "Connection should be TLS")
}

func TestIntegration_ConnectionManager(t *testing.T) {
        server := getTestServer()
        
        config := conn.Config{
                Server:  strings.Split(server, ":")[0],
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        manager := conn.NewManager(config)
        
        // Test connection
        err := manager.Connect()
        if err != nil {
                t.Skipf("Skipping integration test - cannot connect to %s: %v", server, err)
        }
        defer manager.Close()
        
        // Verify connection state
        assert.Equal(t, conn.StateConnected, manager.GetState())
        assert.True(t, manager.IsConnected())
        
        // Test writing through manager
        err = manager.Write("PING :integration_test\r\n")
        require.NoError(t, err)
        
        // Test reading through manager
        reader := manager.Reader()
        assert.NotNil(t, reader)
}

func TestIntegration_IRCHandshake(t *testing.T) {
        server := getTestServer()
        
        config := conn.Config{
                Server:  strings.Split(server, ":")[0],
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        manager := conn.NewManager(config)
        
        // Test connection
        err := manager.Connect()
        if err != nil {
                t.Skipf("Skipping integration test - cannot connect to %s: %v", server, err)
        }
        defer manager.Close()
        
        // Perform IRC handshake
        nick := fmt.Sprintf("testbot_%d", time.Now().Unix())
        
        err = manager.Write(fmt.Sprintf("NICK %s\r\n", nick))
        require.NoError(t, err)
        
        err = manager.Write(fmt.Sprintf("USER %s 0 * :Integration Test Bot\r\n", nick))
        require.NoError(t, err)
        
        // Read server responses
        scanner := bufio.NewScanner(manager.Reader())
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        responses := []string{}
        responseChan := make(chan string, 10)
        
        go func() {
                for scanner.Scan() {
                        line := scanner.Text()
                        select {
                        case responseChan <- line:
                        case <-ctx.Done():
                                return
                        }
                }
        }()
        
        // Collect responses for up to 10 seconds
        timeout := time.After(10 * time.Second)
        for {
                select {
                case line := <-responseChan:
                        responses = append(responses, line)
                        t.Logf("Server: %s", line)
                        
                        // Parse the message
                        msg := parser.ParseMessage(line)
                        
                        // Handle PING
                        if msg.Command == "PING" && len(msg.Params) > 0 {
                                err = manager.Write(fmt.Sprintf("PONG :%s\r\n", msg.Params[0]))
                                require.NoError(t, err)
                        }
                        
                        // Check for successful registration
                        if msg.Command == "001" { // Welcome message
                                t.Log("Successfully registered with IRC server")
                                return
                        }
                        
                        // Check for nick in use
                        if msg.Command == "433" {
                                t.Log("Nick in use, trying alternative")
                                nick = fmt.Sprintf("testbot_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
                                err = manager.Write(fmt.Sprintf("NICK %s\r\n", nick))
                                require.NoError(t, err)
                        }
                        
                case <-timeout:
                        t.Log("Timeout waiting for server responses")
                        goto done
                case <-ctx.Done():
                        goto done
                }
        }
        
done:
        assert.NotEmpty(t, responses, "Should receive responses from server")
        
        // Verify we received some standard IRC responses
        foundPing := false
        foundNumeric := false
        
        for _, response := range responses {
                msg := parser.ParseMessage(response)
                if msg.Command == "PING" {
                        foundPing = true
                }
                if len(msg.Command) == 3 && msg.Command[0] >= '0' && msg.Command[0] <= '9' {
                        foundNumeric = true
                }
        }
        
        if !foundPing && !foundNumeric {
                t.Log("Warning: Did not receive expected IRC protocol messages")
        }
}

func TestIntegration_MessageParsing(t *testing.T) {
        // Test parsing of real IRC messages that might be received
        realMessages := []string{
                ":irc.libera.chat 001 testnick :Welcome to the Libera.Chat IRC Network testnick",
                ":irc.libera.chat 002 testnick :Your host is irc.libera.chat, running version solanum-1.0",
                ":irc.libera.chat 003 testnick :This server was created Mon Jan 1 2024 12:00:00 GMT",
                ":irc.libera.chat 004 testnick irc.libera.chat solanum-1.0 DoQRSXabcdfgiklmnopqrswxyz CFILMPQSTbcefgijklmnopqruvyz bkloveqjfI",
                ":testnick!user@host PRIVMSG #test :Hello, world!",
                ":testnick!user@host JOIN #test",
                ":testnick!user@host PART #test :Leaving",
                ":testnick!user@host QUIT :Connection closed",
                "PING :irc.libera.chat",
                ":irc.libera.chat PONG irc.libera.chat :testnick",
        }
        
        for _, message := range realMessages {
                t.Run(fmt.Sprintf("Parse_%s", strings.Fields(message)[0]), func(t *testing.T) {
                        parsed := parser.ParseMessage(message)
                        
                        // Verify parsing doesn't crash and produces reasonable output
                        assert.NotEmpty(t, parsed.Command, "Command should not be empty")
                        
                        // Verify no dangerous characters in parsed result
                        assert.NotContains(t, parsed.Prefix, "\r")
                        assert.NotContains(t, parsed.Prefix, "\n")
                        assert.NotContains(t, parsed.Command, "\r")
                        assert.NotContains(t, parsed.Command, "\n")
                        
                        for _, param := range parsed.Params {
                                assert.NotContains(t, param, "\r")
                                assert.NotContains(t, param, "\n")
                        }
                        
                        t.Logf("Parsed: Prefix=%q Command=%q Params=%v", 
                                parsed.Prefix, parsed.Command, parsed.Params)
                })
        }
}

func TestIntegration_StressTest(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping stress test in short mode")
        }
        
        server := getTestServer()
        
        config := conn.Config{
                Server:  strings.Split(server, ":")[0],
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        manager := conn.NewManager(config)
        
        // Test connection
        err := manager.Connect()
        if err != nil {
                t.Skipf("Skipping stress test - cannot connect to %s: %v", server, err)
        }
        defer manager.Close()
        
        // Perform rapid writes
        const numWrites = 100
        for i := 0; i < numWrites; i++ {
                message := fmt.Sprintf("PING :stress_test_%d\r\n", i)
                err := manager.Write(message)
                if err != nil {
                        t.Logf("Write failed at iteration %d: %v", i, err)
                        break
                }
                
                // Small delay to avoid overwhelming the server
                time.Sleep(10 * time.Millisecond)
        }
        
        t.Logf("Completed stress test with %d writes", numWrites)
}

// Benchmark real network performance
func BenchmarkIntegration_NetworkWrite(b *testing.B) {
        server := getTestServer()
        
        config := conn.Config{
                Server:  strings.Split(server, ":")[0],
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        manager := conn.NewManager(config)
        
        // Test connection
        err := manager.Connect()
        if err != nil {
                b.Skipf("Skipping benchmark - cannot connect to %s: %v", server, err)
        }
        defer manager.Close()
        
        message := "PING :benchmark_test\r\n"
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
                err := manager.Write(message)
                if err != nil {
                        b.Fatalf("Write failed: %v", err)
                }
        }
}