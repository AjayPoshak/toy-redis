package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"math/rand/v2"

	"github.com/rs/zerolog/log"
)

func PerformGetOperation(connection net.Conn, connectionNumber int) error {
	connection.Write([]byte("GET count\n"))
	buffer := make([]byte, 1024)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Error().Err(err).Int("connection_number", connectionNumber).Msg("Error reading response for GET operation")
		return err
	}
	responseText := strings.TrimSpace(string(buffer[:n]))
	var currentCounter int
	if len(responseText) != 0 {
		currentCounter, err = strconv.Atoi(responseText)
		if err != nil {
			log.Error().Err(err).Int("connection_number", connectionNumber).Msgf("Error converting response to integer for GET operation: %s", responseText)
			return err
		}
	}
	log.Info().Int("connection_number", connectionNumber).Msgf("Current counter value: %d", currentCounter)
	return nil
}

func PerformSetOperation(connection net.Conn, connectionNumber int) error {
	buffer := make([]byte, 1024)
	newCounter := connectionNumber
	connection.Write([]byte("SET count" + fmt.Sprintf(" %d\n", newCounter)))
	_, err := connection.Read(buffer)
	if err != nil {
		log.Error().Err(err).Int("connection_number", connectionNumber).Msgf("Error reading response for SET operation: %s", err)
		return err
	}
	meaningful := strings.Trim(string(buffer), "\x00\n")
	log.Info().Int("connection_number", connectionNumber).Msgf("Write successful new count: %s", meaningful)
	return nil
}

func PerformOperations(connection net.Conn, connectionNumber int) {
	operations := []string{"GET", "GET", "GET", "SET", "SET"}
	for _, operation := range operations {
		switch operation {
		case "GET":
			{
				PerformGetOperation(connection, connectionNumber)
			}
		case "SET":
			{
				PerformSetOperation(connection, connectionNumber)
			}
		}
	}
}

func Connect(connectionNumber int) {
	connection, err := net.Dial("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to server:")
		return
	}
	PerformOperations(connection, connectionNumber)
	defer connection.Close()
}

const TOTAL_CONNECTIONS = 10000
const CONCURRENT_CONNECTIONS = 1000
const HOST = "3.91.71.134"

func SustainedLoadTest() {
	testDuration := 1 * time.Minute
	maxConnections := 1500
	var waitGroup sync.WaitGroup
	var totalOps int64
	var errors int64
	now := time.Now()
	for i := 0; i < maxConnections; i++ {
		waitGroup.Add(1)
		go func(connectionNumber int) {
			connection, err := net.Dial("tcp", HOST+":6379")
			// connection.SetWriteDeadline(time.Now().Add(40 * time.Second))
			// connection.SetReadDeadline(time.Now().Add(40 * time.Second))
			if err != nil {
				atomic.AddInt64(&errors, 1)
				log.Error().Err(err).Int("connection_number", connectionNumber).Msg("Error connecting to server:")
				return
			}
			defer waitGroup.Done()
			defer connection.Close()
			for time.Since(now) < testDuration {
				if rand.IntN(100) < 80 {
					er := PerformGetOperation(connection, connectionNumber)
					if er != nil {
						atomic.AddInt64(&errors, 1)
					}
				} else {
					er := PerformSetOperation(connection, connectionNumber)
					if er != nil {
						atomic.AddInt64(&errors, 1)
					}
				}
				atomic.AddInt64(&totalOps, 1)
				time.Sleep(time.Duration(rand.IntN(100)) * time.Millisecond) // Random sleep to simulate load
			}
		}(i)
	}
	waitGroup.Wait()
	fmt.Println("==== Sustained load test result ====")
	fmt.Println("Total Operations performed:", totalOps)
	fmt.Printf("Total connections: %d\n", maxConnections)
	fmt.Printf("Total errors: %d\n", errors)
	fmt.Printf("Error rate: %.2f%%\n", float64(errors)/float64(totalOps)*100)
	fmt.Printf("Ops per second: %.2f\n", float64(totalOps)/testDuration.Seconds())
}

func main() {
	SustainedLoadTest()
	// var WaitGroup sync.WaitGroup
	// for i := 0; i < TOTAL_CONNECTIONS; i += CONCURRENT_CONNECTIONS {
	// 	WaitGroup.Add(CONCURRENT_CONNECTIONS)
	// 	for j := 0; j < CONCURRENT_CONNECTIONS; j++ {
	// 		go func() {
	// 			Connect(i + j)
	// 			defer WaitGroup.Done()
	// 		}()
	// 	}
	// 	log.Info().Msgf("Waiting for %d connections to finish", CONCURRENT_CONNECTIONS)
	// 	WaitGroup.Wait()
	// 	time.Sleep(5 * time.Second)
	// 	log.Info().Msgf("Finished waiting for %d connections", CONCURRENT_CONNECTIONS)
	// }
}
