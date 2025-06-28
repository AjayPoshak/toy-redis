package redis

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Host string
	Port string
}

type Client struct {
	connection net.Conn
}

var connectionCounter int

func (server *Server) Run() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	log.Info().Msgf("Starting Redis server on %s:%s", server.Host, server.Port)
	storage := NewStore()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.Host, server.Port))
	if err != nil {
		log.Error().Err(err).Msg("Error starting server:")
	}
	defer listener.Close()
	go func() {
		for {
			fmt.Printf("===================> Active goroutines: %d\n", runtime.NumGoroutine())
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		log.Info().Msg("Waiting for new connection...")
		connection, err := listener.Accept()
		log.Info().Msgf("Accepted new connection from %s", connection.RemoteAddr())
		if err != nil {
			log.Error().Err(err).Msg("Error accepting connection:")
			return
		}
		client := &Client{
			connection: connection,
		}
		go client.handleRequest(storage, connectionCounter)
		connectionCounter++
	}
}

func (client *Client) handleRequest(storage *KVStore, connectionCounter int) string {
	reader := bufio.NewReader(client.connection)
	defer client.connection.Close()
	client.connection.SetReadDeadline(time.Now().Add(5 * time.Second))

	for {
		response := ""
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Info().Int("connection_counter", connectionCounter).Msg("Client %d disconnected gracefully")
			} else {
				log.Error().Int("connection_counter", connectionCounter).Err(err).Msg("Error reading from client ")
			}
			return response
		}

		tokens := strings.Split(message, " ")
		if len(tokens) == 0 {
			log.Error().Int("connection_counter", connectionCounter).Msg("Client sent an empty query")
			return ""
		}
		switch tokens[0] {
		case "GET":
			{
				response = GET(storage, tokens, connectionCounter)
				fmt.Println("Response from GET command:", response)
			}

		case "SET":
			{
				log.Info().Int("connection_counter", connectionCounter).Msgf("Client %d sent SET command", connectionCounter)
				response = SET(storage, tokens, connectionCounter)
			}
		}
		log.Info().Int("connection_counter", connectionCounter).Msgf("Sending response back %s", response)
		client.connection.Write([]byte(response))
		client.connection.Write([]byte("\n"))
		log.Info().Int("connection_counter", connectionCounter).Msgf("Response sent to client %s", response)
	}
}
