package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"
	"sync"
	"time"
)

type DataStoreValue struct {
	Value string
	ttl time.Time
}
var DataStore = map[string]DataStoreValue{}
var DataStoreMutex = sync.RWMutex{}

// Simple Strings 
func encodeSimpleStrings(str string) string {
	return "+" + str + "\r\n"
}

func decodeSimpleStrings(cmd string) string {
	return cmd[1 : len(cmd)-6]
}

func decode(msg string) []string {
	if len(msg) == 0 {
		fmt.Println("decode: Failed input message length is 0")
		return []string{}
	}

	if msg[0] != '*' {
		fmt.Println("Invalid input request")
		return []string{}
	}

	index := strings.IndexRune(msg, '\n')
	length, _ := strconv.Atoi(msg[1:index-1])

	result := []string{}
	for i:=0; i<length; i++ {
		index = index+1

		if msg[index] != '$' {
			fmt.Println("Incomplete input request")
			return []string{}
		}

		prev := index
		index = index + strings.IndexRune(msg[index:], '\n')
		size, _ := strconv.Atoi(msg[prev+1:index-1])

		result = append(result, msg[index+1:index+size+1])
		index = index+size+2
	}

	return result
}

func handleCients(conn net.Conn) {
	defer conn.Close()
	for {
		// Read message from connection
		var buffer [512]byte
		length, err := conn.Read(buffer[:])
		if err != nil {
			fmt.Println("Failed to read input" + err.Error())
			break
		}
		msg := string(buffer[:length])

		// Send message to connection
		commands := decode(msg)
		switch strings.ToUpper(commands[0]) {
			case "PING":
				conn.Write([]byte(encodeSimpleStrings("PONG")))
				break
			case "ECHO":
				conn.Write([]byte(encodeSimpleStrings(commands[1])))
				break
			case "GET":
				DataStoreMutex.Lock()
				value, exists := DataStore[commands[1]]
				fmt.Println("ans is", DataStore[commands[1]])

				if !exists {
					conn.Write([]byte("$-1\r\n"));
				} else if (value.ttl != time.Time{} && time.Now().Sub(value.ttl) >= 0) {
					delete(DataStore, commands[1])
					conn.Write([]byte("$-1\r\n"));
				} else {
					conn.Write([]byte(encodeSimpleStrings(value.Value)))
				}

				DataStoreMutex.Unlock()
				break
			case "SET":
				DataStoreMutex.Lock()
				if (len(commands) > 3) {
					expiry, _ := strconv.Atoi(commands[4])
					DataStore[commands[1]] = DataStoreValue{
						Value: commands[2]
						ttl: time.Now().Add(time.Duration(expiry) * time.Millisecond)
					}
				}
				else {
					DataStore[commands[1]] = DataStoreValue{ Value: commands[2] }
				}
				DataStoreMutex.Unlock()
				conn.Write([]byte(encodeSimpleStrings("OK")))
				fmt.Println("ans1 is", DataStore[commands[1]])
				break
			default:
				fmt.Println("handleClients: Something went wrong!!! Didn't Decode")
				break
		}
	}
}

func main() {
	fmt.Println("Your code goes here!")

	// Create a TCP Server
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	} else {
		fmt.Println("Main: Succeedd to bind to port 6379")
	}

	for {
		// Accept the client connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		} else {
			fmt.Println("Main: Succeedd in accepting new connection")
		}

		// Handle multiple clients
		go handleCients(conn)
	}
}
