package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/suryssss/my_redis_go/internal"
)

// handle incoming client requests
func handleConnection(conn net.Conn) {
	defer conn.Close() // close connection when done
	reader := bufio.NewReader(conn)

	// keep reading commands until client disconnects
	for {
		resp, err := internal.Parse(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection:", err)
			}
			break
		}

		reply, err := DispatchCommand(resp)
		if err != nil {
			fmt.Println("Error dispatching command:", err)
			break
		}

		_, err = conn.Write(reply)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			break
		}
	}
}

func main() {
	// start TCP server on port 6379
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	// continuously accept new client connections
	for {
		conn, err := l.Accept() // wait for client
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			break
		}

		// handle each client concurrently
		go handleConnection(conn)
	}
}


