package main

import (
	"fmt"
	"toy-redis/redis"
)

func main() {
	server := &redis.Server{
		Host: "0.0.0.0",
		Port: "6379",
	}
	server.Run()
	fmt.Println("Server running on port 6379")
}
