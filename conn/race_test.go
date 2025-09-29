package conn

import (
        "bufio"
        "fmt"
        "strings"
        "sync"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        "irc-client/testutil"
)

// TestManager_ConcurrentWritesRace tests concurrent write operations for race conditions
func TestManager_ConcurrentWritesRace(t *testing.T) {
        mockConn := testutil.NewMockConn("")
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader("")
        manager.setState(StateConnected)
        
        // Start the writer loop
        go manager.writerLoop()
        
        const numGoroutines = 50
        const writesPerGoroutine = 100
        
        var wg sync.WaitGroup
        errors := make(chan error, numGoroutines*writesPerGoroutine)
        
        // Launch multiple goroutines that write concurrently
        for i := 0; i < numGoroutines; i++ {
                wg.Add(1)
                go func(workerID int) {
                        defer wg.Done()
                        
                        for j := 0; j < writesPerGoroutine; j++ {
                                message := fmt.Sprintf("WORKER_%d_MSG_%d\r\n", workerID, j)
                                if err := manager.Write(message); err != nil {
                                        errors <- err
                                }
                        }
                }(i)
        }
        
        // Wait for all writes to complete
        wg.Wait()
        close(errors)
        
        // Check for any errors
        for err := range errors {
                t.Errorf("Concurrent write error: %v", err)
        }
        
        // Close the manager
        require.NoError(t, manager.Close())
        
        // Verify all messages were written
        written := mockConn.GetWrittenString()
        expectedCount := numGoroutines * writesPerGoroutine
        actualCount := strings.Count(written, "WORKER_")
        
        assert.Equal(t, expectedCount, actualCount, 
                "Expected %d messages, got %d", expectedCount, actualCount)
}

// TestManager_ConcurrentReadWrite tests concurrent read and write operations
func TestManager_ConcurrentReadWrite(t *testing.T) {
        // Prepare mock data for reading
        readData := strings.Repeat("PING :server.example.com\r\n", 100)
        mockConn := testutil.NewMockConn(readData)
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader(readData)
        manager.setState(StateConnected)
        
        // Start the writer loop
        go manager.writerLoop()
        
        var wg sync.WaitGroup
        
        // Start reader goroutine
        wg.Add(1)
        go func() {
                defer wg.Done()
                
                scanner := bufio.NewScanner(manager.Reader())
                for i := 0; i < 50 && scanner.Scan(); i++ {
                        // Simulate processing
                        time.Sleep(time.Millisecond)
                }
        }()
        
        // Start writer goroutine
        wg.Add(1)
        go func() {
                defer wg.Done()
                
                for i := 0; i < 50; i++ {
                        message := fmt.Sprintf("PONG :client_%d\r\n", i)
                        if err := manager.Write(message); err != nil {
                                t.Errorf("Write error: %v", err)
                        }
                        time.Sleep(time.Millisecond)
                }
        }()
        
        // Wait for completion
        wg.Wait()
        
        // Clean shutdown
        require.NoError(t, manager.Close())
}

// TestManager_StateTransitionRaces tests state transitions under concurrent access
func TestManager_StateTransitionRaces(t *testing.T) {
        mockConn := testutil.NewMockConn("")
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader("")
        manager.setState(StateConnected)
        
        go manager.writerLoop()
        
        var wg sync.WaitGroup
        
        // Goroutine 1: Continuously check state
        wg.Add(1)
        go func() {
                defer wg.Done()
                for i := 0; i < 1000; i++ {
                        state := manager.GetState()
                        // State should be valid
                        assert.True(t, state >= StateDisconnected && state <= StateClosed)
                        time.Sleep(time.Microsecond * 10)
                }
        }()
        
        // Goroutine 2: Continuously write
        wg.Add(1)
        go func() {
                defer wg.Done()
                for i := 0; i < 100; i++ {
                        manager.Write(fmt.Sprintf("TEST_%d\r\n", i))
                        time.Sleep(time.Microsecond * 100)
                }
        }()
        
        // Goroutine 3: Check connection status
        wg.Add(1)
        go func() {
                defer wg.Done()
                for i := 0; i < 1000; i++ {
                        manager.IsConnected()
                        time.Sleep(time.Microsecond * 10)
                }
        }()
        
        wg.Wait()
        require.NoError(t, manager.Close())
}

// TestManager_ConcurrentClose tests concurrent close operations
func TestManager_ConcurrentClose(t *testing.T) {
        mockConn := testutil.NewMockConn("")
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader("")
        manager.setState(StateConnected)
        
        go manager.writerLoop()
        
        // Start multiple goroutines that try to close concurrently
        const numClosers = 10
        var wg sync.WaitGroup
        errors := make(chan error, numClosers)
        
        for i := 0; i < numClosers; i++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        err := manager.Close()
                        errors <- err
                }()
        }
        
        wg.Wait()
        close(errors)
        
        // Only the first close should succeed, others should be no-ops
        errorCount := 0
        for err := range errors {
                if err != nil {
                        errorCount++
                }
        }
        
        // All closes should succeed due to sync.Once protection
        assert.Equal(t, 0, errorCount, "All concurrent closes should succeed")
        assert.Equal(t, StateClosed, manager.GetState())
}

// TestManager_WriteAfterCloseRace tests writing after connection is closed
func TestManager_WriteAfterCloseRace(t *testing.T) {
        mockConn := testutil.NewMockConn("")
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader("")
        manager.setState(StateConnected)
        
        go manager.writerLoop()
        
        // Close the manager
        require.NoError(t, manager.Close())
        
        // Try to write after close
        err := manager.Write("TEST\r\n")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "shutting down")
}

// TestManager_RapidConnectDisconnect tests rapid connection state changes
func TestManager_RapidConnectDisconnect(t *testing.T) {
        config := DefaultConfig()
        
        for i := 0; i < 10; i++ {
                mockConn := testutil.NewMockConn("")
                
                manager := NewManager(config)
                manager.conn = mockConn
                writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
                manager.reader = testutil.NewMockBufioReader("")
                manager.setState(StateConnected)
                
                go manager.writerLoop()
                
                // Rapid writes and close
                go func() {
                        for j := 0; j < 5; j++ {
                                manager.Write(fmt.Sprintf("RAPID_%d\r\n", j))
                        }
                }()
                
                time.Sleep(time.Millisecond)
                require.NoError(t, manager.Close())
        }
}

// TestManager_ContextCancellationRace tests context-based cancellation races
func TestManager_ContextCancellationRace(t *testing.T) {
        mockConn := testutil.NewMockConn("")
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader("")
        manager.setState(StateConnected)
        
        go manager.writerLoop()
        
        var wg sync.WaitGroup
        
        // Start writes that should be interrupted by context cancellation
        wg.Add(1)
        go func() {
                defer wg.Done()
                
                for i := 0; i < 100; i++ {
                        select {
                        case <-manager.Context().Done():
                                return
                        default:
                                manager.Write(fmt.Sprintf("CONTEXT_TEST_%d\r\n", i))
                                time.Sleep(time.Millisecond)
                        }
                }
        }()
        
        // Cancel after short delay
        time.Sleep(10 * time.Millisecond)
        manager.Close()
        
        wg.Wait()
        
        // Context should be done
        select {
        case <-manager.Context().Done():
                // Expected
        default:
                t.Error("Context should be cancelled after Close()")
        }
}

// Benchmark concurrent operations
func BenchmarkManager_ConcurrentWrites(b *testing.B) {
        mockConn := testutil.NewMockConn("")
        
        config := DefaultConfig()
        manager := NewManager(config)
        manager.conn = mockConn
        writer, _ := testutil.NewMockBufioWriter()
        manager.writer = writer
        manager.reader = testutil.NewMockBufioReader("")
        manager.setState(StateConnected)
        
        go manager.writerLoop()
        defer manager.Close()
        
        b.ResetTimer()
        b.RunParallel(func(pb *testing.PB) {
                i := 0
                for pb.Next() {
                        manager.Write(fmt.Sprintf("BENCH_%d\r\n", i))
                        i++
                }
        })
}