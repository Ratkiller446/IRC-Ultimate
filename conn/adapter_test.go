package conn

import (
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"

        "irc-client/internal/app"
)

func TestNewAdapter(t *testing.T) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        adapter := NewAdapter(config)
        require.NotNil(t, adapter)
        
        // Should implement app.Connection interface
        var _ app.Connection = adapter
        
        // Should be able to cast back to get manager
        concreteAdapter, ok := adapter.(*Adapter)
        require.True(t, ok)
        require.NotNil(t, concreteAdapter.GetManager())
}

func TestNewAdapterFromManager(t *testing.T) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        manager := NewManager(config)
        adapter := NewAdapterFromManager(manager)
        require.NotNil(t, adapter)
        
        // Should implement app.Connection interface
        var _ app.Connection = adapter
        
        // Should wrap the same manager
        concreteAdapter := adapter.(*Adapter)
        assert.Equal(t, manager, concreteAdapter.GetManager())
}

func TestAdapter_MethodDelegation(t *testing.T) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 10 * time.Second,
        }
        
        adapter := NewAdapter(config).(*Adapter)
        manager := adapter.GetManager()
        
        // Test that methods are properly delegated
        
        // Test Context
        adapterCtx := adapter.Context()
        managerCtx := manager.Context()
        assert.Equal(t, managerCtx, adapterCtx)
        
        // Test IsConnected (should start false)
        assert.Equal(t, manager.IsConnected(), adapter.IsConnected())
        assert.False(t, adapter.IsConnected())
        
        // Test GetState (accessed through concrete adapter)
        concreteAdapter := adapter.(*Adapter)
        assert.Equal(t, manager.GetState(), concreteAdapter.GetState())
        assert.Equal(t, StateDisconnected, concreteAdapter.GetState())
        
        // Test ErrorChannel
        adapterErrCh := adapter.ErrorChannel()
        managerErrCh := manager.ErrorChannel()
        assert.Equal(t, managerErrCh, adapterErrCh)
        
        // Test Reader (should not be nil even when disconnected)
        assert.NotNil(t, adapter.Reader())
        assert.Equal(t, manager.Reader(), adapter.Reader())
}

func TestAdapter_ConnectAndClose(t *testing.T) {
        // Use a mock connection for testing without actual network
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config)
        
        // Test Connect error (should fail to connect to fake server)
        err := adapter.Connect()
        assert.Error(t, err) // Expected to fail with fake server
        
        // Test Close (should not error even if not connected)
        err = adapter.Close()
        assert.NoError(t, err)
}

func TestAdapter_WriteWithoutConnection(t *testing.T) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config)
        
        // Test Write without connection (should error)
        err := adapter.Write("TEST MESSAGE")
        assert.Error(t, err)
}

func TestAdapter_IntegrationWithMockConnection(t *testing.T) {
        // Create adapter
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config).(*Adapter)
        
        // Test integration points
        assert.NotNil(t, adapter.GetManager())
        assert.Equal(t, StateDisconnected, adapter.GetState())
        
        // Test that context is valid
        ctx := adapter.Context()
        assert.NotNil(t, ctx)
        
        // Test that context is cancelled when closed
        go func() {
                time.Sleep(10 * time.Millisecond)
                adapter.Close()
        }()
        
        select {
        case <-ctx.Done():
                // Context was properly cancelled
        case <-time.After(100 * time.Millisecond):
                t.Error("Context was not cancelled within expected time")
        }
}

func TestAdapter_ErrorChannelHandling(t *testing.T) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config)
        
        // Get error channel
        errCh := adapter.ErrorChannel()
        assert.NotNil(t, errCh)
        
        // Close adapter
        err := adapter.Close()
        assert.NoError(t, err)
        
        // Error channel should be closed
        select {
        case _, ok := <-errCh:
                assert.False(t, ok, "Error channel should be closed")
        case <-time.After(100 * time.Millisecond):
                t.Error("Error channel was not closed within expected time")
        }
}

// Test concurrent usage
func TestAdapter_ConcurrentAccess(t *testing.T) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config)
        
        // Start multiple goroutines accessing the adapter
        done := make(chan bool, 3)
        
        // Goroutine 1: Check connection state
        go func() {
                for i := 0; i < 100; i++ {
                        adapter.IsConnected()
                        adapter.GetState()
                }
                done <- true
        }()
        
        // Goroutine 2: Access reader and context
        go func() {
                for i := 0; i < 100; i++ {
                        adapter.Reader()
                        adapter.Context()
                }
                done <- true
        }()
        
        // Goroutine 3: Access error channel
        go func() {
                for i := 0; i < 100; i++ {
                        adapter.ErrorChannel()
                }
                done <- true
        }()
        
        // Wait for all goroutines to complete
        for i := 0; i < 3; i++ {
                select {
                case <-done:
                        // Goroutine completed successfully
                case <-time.After(5 * time.Second):
                        t.Fatal("Goroutine did not complete within expected time")
                }
        }
        
        // Clean up
        adapter.Close()
}

// Benchmark tests
func BenchmarkAdapter_IsConnected(b *testing.B) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config)
        defer adapter.Close()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
                adapter.IsConnected()
        }
}

func BenchmarkAdapter_GetState(b *testing.B) {
        config := Config{
                Server:  "irc.example.com",
                Port:    6667,
                TLS:     false,
                Timeout: 1 * time.Second,
        }
        
        adapter := NewAdapter(config)
        defer adapter.Close()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
                adapter.GetState()
        }
}

// Test that Adapter properly implements app.Connection interface
func TestAdapter_ImplementsInterface(t *testing.T) {
        var _ app.Connection = (*Adapter)(nil)
}