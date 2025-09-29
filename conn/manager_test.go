package conn

import (
        "bufio"
        "fmt"
        "net"
        "strings"
        "sync"
        "testing"
        "time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
        readData  []byte
        writeData []byte
        closed    bool
        mu        sync.RWMutex
        readPos   int
}

func newMockConn(readData string) *mockConn {
        return &mockConn{
                readData: []byte(readData),
        }
}

func (m *mockConn) Read(b []byte) (int, error) {
        m.mu.RLock()
        defer m.mu.RUnlock()
        
        if m.closed {
                return 0, fmt.Errorf("connection closed")
        }
        
        if m.readPos >= len(m.readData) {
                // Simulate blocking read
                time.Sleep(100 * time.Millisecond)
                return 0, fmt.Errorf("EOF")
        }
        
        n := copy(b, m.readData[m.readPos:])
        m.readPos += n
        return n, nil
}

func (m *mockConn) Write(b []byte) (int, error) {
        m.mu.Lock()
        defer m.mu.Unlock()
        
        if m.closed {
                return 0, fmt.Errorf("connection closed")
        }
        
        m.writeData = append(m.writeData, b...)
        return len(b), nil
}

func (m *mockConn) Close() error {
        m.mu.Lock()
        defer m.mu.Unlock()
        m.closed = true
        return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *mockConn) GetWrittenData() string {
        m.mu.RLock()
        defer m.mu.RUnlock()
        return string(m.writeData)
}

func (m *mockConn) IsClosed() bool {
        m.mu.RLock()
        defer m.mu.RUnlock()
        return m.closed
}

// TestManager_SingleWriterPattern tests that only one goroutine writes at a time
func TestManager_SingleWriterPattern(t *testing.T) {
        tests := []struct {
                name        string
                numWriters  int
                writesEach  int
                expectRace  bool
        }{
                {
                        name:       "single writer",
                        numWriters: 1,
                        writesEach: 10,
                        expectRace: false,
                },
                {
                        name:       "multiple writers",
                        numWriters: 10,
                        writesEach: 10,
                        expectRace: false, // Should not race due to single-writer pattern
                },
                {
                        name:       "high concurrency",
                        numWriters: 50,
                        writesEach: 20,
                        expectRace: false, // Should not race due to single-writer pattern
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        // Create mock connection
                        mockConn := newMockConn("")
                        
                        // Create manager with mock connection
                        config := DefaultConfig()
                        manager := NewManager(config)
                        manager.conn = mockConn
                        manager.writer = bufio.NewWriter(mockConn)
                        manager.reader = bufio.NewReader(mockConn)
                        manager.setState(StateConnected)
                        
                        // Start the writer loop
                        go manager.writerLoop()
                        
                        var wg sync.WaitGroup
                        
                        // Start multiple writer goroutines
                        for i := 0; i < tt.numWriters; i++ {
                                wg.Add(1)
                                go func(writerID int) {
                                        defer wg.Done()
                                        for j := 0; j < tt.writesEach; j++ {
                                                message := fmt.Sprintf("WRITER_%d_MSG_%d\r\n", writerID, j)
                                                if err := manager.Write(message); err != nil {
                                                        t.Errorf("Writer %d failed to write message %d: %v", writerID, j, err)
                                                }
                                        }
                                }(i)
                        }
                        
                        // Wait for all writers to complete
                        wg.Wait()
                        
                        // Close the manager
                        if err := manager.Close(); err != nil {
                                t.Errorf("Failed to close manager: %v", err)
                        }
                        
                        // Verify all writes were successful
                        written := mockConn.GetWrittenData()
                        expectedWrites := tt.numWriters * tt.writesEach
                        actualWrites := strings.Count(written, "WRITER_")
                        
                        if actualWrites != expectedWrites {
                                t.Errorf("Expected %d writes, got %d", expectedWrites, actualWrites)
                        }
                })
        }
}

// TestManager_StateTransitions tests connection state management
func TestManager_StateTransitions(t *testing.T) {
        tests := []struct {
                name           string
                initialState   ConnectionState
                operation      func(*Manager) error
                expectedState  ConnectionState
                expectError    bool
        }{
                {
                        name:         "connect from disconnected",
                        initialState: StateDisconnected,
                        operation: func(m *Manager) error {
                                // Mock the connection process
                                m.conn = newMockConn("")
                                m.writer = bufio.NewWriter(m.conn)
                                m.reader = bufio.NewReader(m.conn)
                                m.setState(StateConnected)
                                go m.writerLoop()
                                return nil
                        },
                        expectedState: StateConnected,
                        expectError:   false,
                },
                {
                        name:         "close from connected",
                        initialState: StateConnected,
                        operation: func(m *Manager) error {
                                return m.Close()
                        },
                        expectedState: StateClosed,
                        expectError:   false,
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        config := DefaultConfig()
                        manager := NewManager(config)
                        manager.setState(tt.initialState)
                        
                        err := tt.operation(manager)
                        
                        if tt.expectError && err == nil {
                                t.Error("Expected error but got none")
                        }
                        if !tt.expectError && err != nil {
                                t.Errorf("Unexpected error: %v", err)
                        }
                        
                        if manager.GetState() != tt.expectedState {
                                t.Errorf("Expected state %v, got %v", tt.expectedState, manager.GetState())
                        }
                })
        }
}

// TestManager_WriteTimeout tests write timeout functionality
func TestManager_WriteTimeout(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        manager.writeTimeout = 100 * time.Millisecond
        
        // Set up mock connection that blocks on write
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Don't start writer loop to simulate blocked writer
        
        // Attempt write - should timeout
        start := time.Now()
        err := manager.Write("TEST MESSAGE\r\n")
        duration := time.Since(start)
        
        if err == nil {
                t.Error("Expected timeout error but got none")
        }
        
        if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
                t.Errorf("Expected timeout around 100ms, got %v", duration)
        }
}

// TestManager_ContextCancellation tests context-based cancellation
func TestManager_ContextCancellation(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Set up mock connection
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Start writer loop
        go manager.writerLoop()
        
        // Cancel context
        manager.cancel()
        
        // Wait a bit for cancellation to propagate
        time.Sleep(50 * time.Millisecond)
        
        // Attempt write - should fail due to cancelled context
        err := manager.Write("TEST MESSAGE\r\n")
        if err == nil {
                t.Error("Expected error due to cancelled context but got none")
        }
        
        if !strings.Contains(err.Error(), "shutting down") {
                t.Errorf("Expected 'shutting down' error, got: %v", err)
        }
}

// TestManager_ErrorHandling tests error propagation
func TestManager_ErrorHandling(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Test error reporting
        testError := fmt.Errorf("test error")
        manager.reportError(testError)
        
        select {
        case err := <-manager.ErrorChannel():
                if err.Error() != "test error" {
                        t.Errorf("Expected 'test error', got: %v", err)
                }
        case <-time.After(100 * time.Millisecond):
                t.Error("Expected error on error channel but got timeout")
        }
}

// TestManager_ConcurrentWrites tests race conditions in write operations
func TestManager_ConcurrentWrites(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Set up mock connection
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Start writer loop
        go manager.writerLoop()
        
        const numGoroutines = 100
        const messagesPerGoroutine = 10
        
        var wg sync.WaitGroup
        errors := make(chan error, numGoroutines*messagesPerGoroutine)
        
        // Start concurrent writers
        for i := 0; i < numGoroutines; i++ {
                wg.Add(1)
                go func(goroutineID int) {
                        defer wg.Done()
                        for j := 0; j < messagesPerGoroutine; j++ {
                                message := fmt.Sprintf("GOROUTINE_%d_MSG_%d\r\n", goroutineID, j)
                                if err := manager.Write(message); err != nil {
                                        errors <- err
                                }
                        }
                }(i)
        }
        
        wg.Wait()
        close(errors)
        
        // Check for any errors
        var errorCount int
        for err := range errors {
                t.Errorf("Concurrent write error: %v", err)
                errorCount++
        }
        
        if errorCount > 0 {
                t.Errorf("Got %d errors during concurrent writes", errorCount)
        }
        
        // Verify all messages were written
        written := mockConn.GetWrittenData()
        expectedCount := numGoroutines * messagesPerGoroutine
        actualCount := strings.Count(written, "GOROUTINE_")
        
        if actualCount != expectedCount {
                t.Errorf("Expected %d messages, got %d", expectedCount, actualCount)
        }
        
        // Close manager
        if err := manager.Close(); err != nil {
                t.Errorf("Failed to close manager: %v", err)
        }
}

// BenchmarkManager_Write benchmarks write performance
func BenchmarkManager_Write(b *testing.B) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Set up mock connection
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Start writer loop
        go manager.writerLoop()
        
        b.ResetTimer()
        
        for i := 0; i < b.N; i++ {
                if err := manager.Write("BENCHMARK MESSAGE\r\n"); err != nil {
                        b.Fatalf("Write failed: %v", err)
                }
        }
        
        manager.Close()
}

// TestManager_ConcurrentCloseWrite tests the critical race condition between Close() and Write()
func TestManager_ConcurrentCloseWrite(t *testing.T) {
        tests := []struct {
                name            string
                numWriters      int
                numCloseAttempts int
                testDuration    time.Duration
        }{
                {
                        name:            "single writer with close",
                        numWriters:      1,
                        numCloseAttempts: 1,
                        testDuration:    100 * time.Millisecond,
                },
                {
                        name:            "multiple writers with close",
                        numWriters:      10,
                        numCloseAttempts: 1,
                        testDuration:    200 * time.Millisecond,
                },
                {
                        name:            "high concurrency stress test",
                        numWriters:      50,
                        numCloseAttempts: 3,
                        testDuration:    500 * time.Millisecond,
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        // Run multiple iterations to increase chance of hitting race condition
                        for i := 0; i < 10; i++ {
                                func() {
                                        config := DefaultConfig()
                                        manager := NewManager(config)
                                        
                                        // Set up mock connection
                                        mockConn := newMockConn("")
                                        manager.conn = mockConn
                                        manager.writer = bufio.NewWriter(mockConn)
                                        manager.reader = bufio.NewReader(mockConn)
                                        manager.setState(StateConnected)
                                        
                                        // Start writer loop
                                        go manager.writerLoop()
                                        
                                        var wg sync.WaitGroup
                                        panicChan := make(chan interface{}, tt.numWriters)
                                        
                                        // Start writers that will race with Close()
                                        for j := 0; j < tt.numWriters; j++ {
                                                wg.Add(1)
                                                go func(writerID int) {
                                                        defer wg.Done()
                                                        defer func() {
                                                                if r := recover(); r != nil {
                                                                        panicChan <- r
                                                                }
                                                        }()
                                                        
                                                        start := time.Now()
                                                        msgCount := 0
                                                        for time.Since(start) < tt.testDuration {
                                                                message := fmt.Sprintf("WRITER_%d_MSG_%d\r\n", writerID, msgCount)
                                                                // This should never panic, even if Close() is called concurrently
                                                                manager.Write(message)
                                                                msgCount++
                                                                // Small delay to allow Close() to potentially race
                                                                time.Sleep(time.Microsecond)
                                                        }
                                                }(j)
                                        }
                                        
                                        // Start closers that will race with Write()
                                        for j := 0; j < tt.numCloseAttempts; j++ {
                                                wg.Add(1)
                                                go func(closeID int) {
                                                        defer wg.Done()
                                                        time.Sleep(time.Duration(closeID+1) * 50 * time.Millisecond)
                                                        manager.Close()
                                                }(j)
                                        }
                                        
                                        wg.Wait()
                                        close(panicChan)
                                        
                                        // Check for any panics
                                        var panics []interface{}
                                        for p := range panicChan {
                                                panics = append(panics, p)
                                        }
                                        
                                        if len(panics) > 0 {
                                                t.Errorf("Iteration %d: Got %d panics during concurrent Close()+Write(): %v", i, len(panics), panics)
                                        }
                                        
                                        // Ensure final cleanup
                                        manager.Close()
                                }()
                        }
                })
        }
}

// TestManager_ContextCancellationDuringConnect tests context cancellation during connection attempts
func TestManager_ContextCancellationDuringConnect(t *testing.T) {
        config := DefaultConfig()
        config.Server = "nonexistent.invalid.test"
        config.Port = 99999
        config.Timeout = 10 * time.Second // Long timeout to ensure we can cancel first
        
        manager := NewManager(config)
        
        // Start connection attempt in background
        connectDone := make(chan error, 1)
        go func() {
                connectDone <- manager.Connect()
        }()
        
        // Cancel context after short delay
        time.Sleep(50 * time.Millisecond)
        manager.Close()
        
        // Connection should fail due to cancellation, not timeout
        select {
        case err := <-connectDone:
                if err == nil {
                        t.Error("Expected connection to fail due to cancellation")
                }
                // Error should occur quickly due to context cancellation
        case <-time.After(2 * time.Second):
                t.Error("Connection attempt did not respect context cancellation - took too long")
        }
}

// TestManager_ShutdownWithPendingWrites tests shutdown behavior with pending writes
func TestManager_ShutdownWithPendingWrites(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Set up mock connection
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Start writer loop
        go manager.writerLoop()
        
        // Queue up several writes
        numWrites := 10
        var wg sync.WaitGroup
        errors := make(chan error, numWrites)
        
        for i := 0; i < numWrites; i++ {
                wg.Add(1)
                go func(id int) {
                        defer wg.Done()
                        message := fmt.Sprintf("PENDING_WRITE_%d\r\n", id)
                        err := manager.Write(message)
                        if err != nil {
                                errors <- err
                        }
                }(i)
        }
        
        // Close manager while writes are in progress
        time.Sleep(10 * time.Millisecond)
        closeErr := manager.Close()
        
        // Wait for all writes to complete
        wg.Wait()
        close(errors)
        
        // Close should succeed
        if closeErr != nil {
                t.Errorf("Close failed: %v", closeErr)
        }
        
        // Some writes may fail due to shutdown, which is acceptable
        var errorCount int
        for err := range errors {
                if !strings.Contains(err.Error(), "shutting down") && 
                   !strings.Contains(err.Error(), "not established") {
                        t.Errorf("Unexpected error during shutdown: %v", err)
                }
                errorCount++
        }
        
        t.Logf("Got %d write errors during shutdown (expected)", errorCount)
}

// TestManager_RepeatedCloseOperations tests that repeated Close() calls are safe
func TestManager_RepeatedCloseOperations(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Set up mock connection
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Start writer loop
        go manager.writerLoop()
        
        // Call Close() multiple times concurrently
        var wg sync.WaitGroup
        numClosers := 10
        errors := make(chan error, numClosers)
        
        for i := 0; i < numClosers; i++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        err := manager.Close()
                        if err != nil {
                                errors <- err
                        }
                }()
        }
        
        wg.Wait()
        close(errors)
        
        // Only the first Close() should potentially return an error from closing the connection
        // All subsequent calls should be no-ops
        var errorCount int
        for err := range errors {
                errorCount++
                t.Logf("Close error: %v", err)
        }
        
        // At most one error is acceptable (from the actual connection close)
        if errorCount > 1 {
                t.Errorf("Too many close errors: %d", errorCount)
        }
        
        // Manager should be in closed state
        if manager.GetState() != StateClosed {
                t.Errorf("Expected state %v, got %v", StateClosed, manager.GetState())
        }
}

// TestManager_WriteAfterClose tests that writes after close fail gracefully
func TestManager_WriteAfterClose(t *testing.T) {
        config := DefaultConfig()
        manager := NewManager(config)
        
        // Set up mock connection
        mockConn := newMockConn("")
        manager.conn = mockConn
        manager.writer = bufio.NewWriter(mockConn)
        manager.reader = bufio.NewReader(mockConn)
        manager.setState(StateConnected)
        
        // Start writer loop
        go manager.writerLoop()
        
        // Close the manager
        if err := manager.Close(); err != nil {
                t.Errorf("Close failed: %v", err)
        }
        
        // Wait for shutdown to complete
        time.Sleep(50 * time.Millisecond)
        
        // Attempts to write should fail gracefully
        for i := 0; i < 5; i++ {
                err := manager.Write(fmt.Sprintf("POST_CLOSE_MSG_%d\r\n", i))
                if err == nil {
                        t.Error("Expected write to fail after close")
                }
                if !strings.Contains(err.Error(), "shutting down") &&
                   !strings.Contains(err.Error(), "not established") {
                        t.Errorf("Unexpected error message: %v", err)
                }
        }
}

// TestConfig_Validation tests configuration validation
func TestConfig_Validation(t *testing.T) {
        tests := []struct {
                name        string
                config      Config
                expectError bool
                errorMsg    string
        }{
                {
                        name:        "valid config",
                        config:      DefaultConfig(),
                        expectError: false,
                },
                {
                        name: "empty server",
                        config: Config{
                                Server:  "",
                                Port:    6667,
                                Timeout: 30 * time.Second,
                        },
                        expectError: true,
                        errorMsg:    "server address is required",
                },
                {
                        name: "invalid port",
                        config: Config{
                                Server:  "irc.example.com",
                                Port:    -1,
                                Timeout: 30 * time.Second,
                        },
                        expectError: true,
                        errorMsg:    "invalid port number",
                },
                {
                        name: "zero timeout",
                        config: Config{
                                Server:  "irc.example.com",
                                Port:    6667,
                                Timeout: 0,
                        },
                        expectError: true,
                        errorMsg:    "timeout must be positive",
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        err := ValidateConfig(tt.config)
                        
                        if tt.expectError && err == nil {
                                t.Error("Expected validation error but got none")
                        }
                        
                        if !tt.expectError && err != nil {
                                t.Errorf("Unexpected validation error: %v", err)
                        }
                        
                        if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
                                t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMsg, err)
                        }
                })
        }
}