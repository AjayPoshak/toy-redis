package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"toy-redis/redis"
)

var portCounter int64 = 6400

// getNextPort returns a unique port for each test
func getNextPort() string {
	return fmt.Sprintf("%d", atomic.AddInt64(&portCounter, 1))
}

// TestClient represents a test client connection
type TestClient struct {
	conn   net.Conn
	reader *bufio.Reader
}

// NewTestClient creates a new test client and connects to the server
func NewTestClient(t *testing.T, port string) *TestClient {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		t.Fatalf("Failed to connect to server on port %s: %v", port, err)
	}
	
	return &TestClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// SendCommand sends a command to the server and returns the response
func (tc *TestClient) SendCommand(command string) (string, error) {
	// Send command
	_, err := tc.conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}
	
	// Read response
	response, err := tc.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	
	return strings.TrimSpace(response), nil
}

// Close closes the client connection
func (tc *TestClient) Close() {
	tc.conn.Close()
}

// startTestServer starts the server in a goroutine for testing
func startTestServer(t *testing.T, port string) {
	server := &redis.Server{
		Host: "localhost",
		Port: port,
	}
	
	go func() {
		server.Run()
	}()
	
	// Wait for server to start with retries
	maxRetries := 20
	for i := 0; i < maxRetries; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	
	t.Fatalf("Server failed to start on port %s within expected time", port)
}

func TestBasicSetAndGet(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	client := NewTestClient(t, port)
	defer client.Close()
	
	// Test SET command
	response, err := client.SendCommand("SET mykey myvalue")
	if err != nil {
		t.Fatalf("SET command failed: %v", err)
	}
	if response != "myvalue" {
		t.Errorf("Expected SET response 'myvalue', got '%s'", response)
	}
	
	// Test GET command
	response, err = client.SendCommand("GET mykey")
	if err != nil {
		t.Fatalf("GET command failed: %v", err)
	}
	if response != "myvalue" {
		t.Errorf("Expected GET response 'myvalue', got '%s'", response)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	client := NewTestClient(t, port)
	defer client.Close()
	
	// Test GET for non-existent key
	response, err := client.SendCommand("GET nonexistent")
	if err != nil {
		t.Fatalf("GET command failed: %v", err)
	}
	if response != "" {
		t.Errorf("Expected empty response for non-existent key, got '%s'", response)
	}
}

func TestOverwriteKey(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	client := NewTestClient(t, port)
	defer client.Close()
	
	// Set initial value
	_, err := client.SendCommand("SET testkey value1")
	if err != nil {
		t.Fatalf("First SET command failed: %v", err)
	}
	
	// Overwrite with new value
	response, err := client.SendCommand("SET testkey value2")
	if err != nil {
		t.Fatalf("Second SET command failed: %v", err)
	}
	if response != "value2" {
		t.Errorf("Expected SET response 'value2', got '%s'", response)
	}
	
	// Verify new value
	response, err = client.SendCommand("GET testkey")
	if err != nil {
		t.Fatalf("GET command failed: %v", err)
	}
	if response != "value2" {
		t.Errorf("Expected GET response 'value2', got '%s'", response)
	}
}

func TestMultipleKeysAndValues(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	client := NewTestClient(t, port)
	defer client.Close()
	
	testCases := map[string]string{
		"key1":    "value1",
		"key2":    "value2", 
		"longkey": "verylongvalue",
		"numkey":  "12345",
		// Note: Values with spaces are not supported by current server implementation
		"underscorekey": "value_with_underscores",
	}
	
	// Set all keys
	for key, value := range testCases {
		response, err := client.SendCommand(fmt.Sprintf("SET %s %s", key, value))
		if err != nil {
			t.Fatalf("SET command failed for key %s: %v", key, err)
		}
		if response != value {
			t.Errorf("Expected SET response '%s' for key %s, got '%s'", value, key, response)
		}
	}
	
	// Get all keys and verify values
	for key, expectedValue := range testCases {
		response, err := client.SendCommand(fmt.Sprintf("GET %s", key))
		if err != nil {
			t.Fatalf("GET command failed for key %s: %v", key, err)
		}
		if response != expectedValue {
			t.Errorf("Expected GET response '%s' for key %s, got '%s'", expectedValue, key, response)
		}
	}
}

func TestConcurrentClients(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	numClients := 10
	numOperations := 20
	var wg sync.WaitGroup
	errors := make(chan error, numClients*numOperations)
	
	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			client := NewTestClient(t, port)
			defer client.Close()
			
			for op := 0; op < numOperations; op++ {
				key := fmt.Sprintf("client%d_key%d", id, op)
				value := fmt.Sprintf("client%d_value%d", id, op)
				
				// SET operation
				response, err := client.SendCommand(fmt.Sprintf("SET %s %s", key, value))
				if err != nil {
					errors <- fmt.Errorf("client %d SET failed: %v", id, err)
					return
				}
				if response != value {
					errors <- fmt.Errorf("client %d SET response mismatch: expected %s, got %s", id, value, response)
					return
				}
				
				// GET operation
				response, err = client.SendCommand(fmt.Sprintf("GET %s", key))
				if err != nil {
					errors <- fmt.Errorf("client %d GET failed: %v", id, err)
					return
				}
				if response != value {
					errors <- fmt.Errorf("client %d GET response mismatch: expected %s, got %s", id, value, response)
					return
				}
			}
		}(clientID)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	// Pre-populate some data
	setupClient := NewTestClient(t, port)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("shared_key_%d", i)
		value := fmt.Sprintf("initial_value_%d", i)
		setupClient.SendCommand(fmt.Sprintf("SET %s %s", key, value))
	}
	setupClient.Close()
	
	numReaders := 5
	numWriters := 3
	duration := 2 * time.Second
	var wg sync.WaitGroup
	errors := make(chan error, numReaders+numWriters)
	
	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			
			client := NewTestClient(t, port)
			defer client.Close()
			
			start := time.Now()
			for time.Since(start) < duration {
				key := fmt.Sprintf("shared_key_%d", readerID%10)
				_, err := client.SendCommand(fmt.Sprintf("GET %s", key))
				if err != nil {
					errors <- fmt.Errorf("reader %d failed: %v", readerID, err)
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}
	
	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			
			client := NewTestClient(t, port)
			defer client.Close()
			
			start := time.Now()
			counter := 0
			for time.Since(start) < duration {
				key := fmt.Sprintf("shared_key_%d", writerID%10)
				value := fmt.Sprintf("writer_%d_value_%d", writerID, counter)
				_, err := client.SendCommand(fmt.Sprintf("SET %s %s", key, value))
				if err != nil {
					errors <- fmt.Errorf("writer %d failed: %v", writerID, err)
					return
				}
				counter++
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

func TestPerformanceMetrics(t *testing.T) {
	port := getNextPort()
	startTestServer(t, port)
	
	numClients := 10
	testDuration := 3 * time.Second
	var totalOps int64
	var wg sync.WaitGroup
	
	start := time.Now()
	
	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			client := NewTestClient(t, port)
			defer client.Close()
			
			ops := 0
			clientStart := time.Now()
			
			for time.Since(clientStart) < testDuration {
				key := fmt.Sprintf("perf_client_%d_key_%d", id, ops)
				value := fmt.Sprintf("perf_client_%d_value_%d", id, ops)
				
				// SET operation
				_, err := client.SendCommand(fmt.Sprintf("SET %s %s", key, value))
				if err != nil {
					t.Errorf("Client %d SET failed: %v", id, err)
					return
				}
				
				// GET operation
				_, err = client.SendCommand(fmt.Sprintf("GET %s", key))
				if err != nil {
					t.Errorf("Client %d GET failed: %v", id, err)
					return
				}
				
				ops += 2 // Count both SET and GET
			}
			
			atomic.AddInt64(&totalOps, int64(ops))
		}(clientID)
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	opsPerSecond := float64(totalOps) / elapsed.Seconds()
	
	t.Logf("Performance Test Results:")
	t.Logf("Total Operations: %d", totalOps)
	t.Logf("Test Duration: %v", elapsed)
	t.Logf("Operations per Second: %.2f", opsPerSecond)
	t.Logf("Concurrent Clients: %d", numClients)
	
	// Basic performance assertion - should handle at least 100 ops/sec
	if opsPerSecond < 100 {
		t.Errorf("Performance below threshold: %.2f ops/sec (expected > 100)", opsPerSecond)
	}
}
