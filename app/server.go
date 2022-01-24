package main

import (
	"bufio" // Read Input
	"fmt"   // Send Message
	"net"   // Connection Socket
	"os"    // Control Program
)

func main() {
	fmt.Println("Your code goes here!")

	// Create a TCP Server
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	// Accept the client connection
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// Chat
	for {
		// Read client commands
		_, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Print(err.Error())
		}

		// Respond to client
		conn.Write([]byte("+PONG\r\n"))
	}
}
