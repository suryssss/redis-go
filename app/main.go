package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// handle incoming client requests
func handleConnection(conn net.Conn) {
	defer conn.Close() // close connection when done

	buf := make([]byte, 1024) // buffer to store client data

	// keep reading data until client disconnects
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}

		data := string(buf[:n])
		fmt.Println("Received data:", data)

		// check for PING command and reply with PONG
		if strings.Contains(data, "PING") {
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				break
			}
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

	// continuously accept new client connections
	for {
		conn, err := l.Accept() // wait for client
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}

		// handle each client concurrently
		go handleConnection(conn)
	}
}

