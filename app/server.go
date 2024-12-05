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

/////////////////////////////////////////////
// Simple Strings 
func encodeSimpleStrings(str string) string {
	return "+" + str + "\r\n"
}

func decodeSimpleStrings(cmd string) string {
	return cmd[1 : len(cmd)-6]
}

// Bulk Strings 
func encodeBulkStrings(str string) string {
	return "$" + strconv.Itoa(len(str)) + "\r\n" + str + "\r\n"
}

func decodeBulkStrings(cmd string) string {
	for index, char := range cmd[1:] {
		if !unicode.IsNumber(char) {
			cmd = cmd[index:]
			break
		}
	}
	return cmd[5 : len(cmd)-6]
}

// Arrays
func encodeArrays(str string) string {
	result := "*" + string(rune(len(str))) + "\r\n"

	list := strings.Split(str, " ")
	for _, word := range list {
		result += encodeBulkStrings(word)
	}

	return result
}

func decodeArrays(cmd string) string {
	result := ""
	var prevIndex int = 6
	var prevChar string = "$"
	var test string
	for i := 0; i < len(cmd)-1; i++ {
		if (cmd[i:i+2] != "\n") && (cmd[i:i+2] != "\r") {
			test += string(cmd[i])
		}
	}

	for index, letter := range cmd[7:] {
		if letter == '$' {
			result += decodeBulkStrings(cmd[prevIndex:index])
			prevIndex = index + 1
			prevChar = "$"
		} else if letter == '+' {
			result += decodeSimpleStrings(cmd[prevIndex:index])
			prevIndex = index + 1
			prevChar = "+"
		} else if letter == '*' {
			result += decodeArrays(cmd[prevIndex:index])
			prevIndex = index + 1
			prevChar = "*"
		}
		if index == len(cmd[7:])-1 {
			if prevChar == "$" {
				result += decodeBulkStrings(cmd[prevIndex:])
			} else if prevChar == "+" {
				result += decodeSimpleStrings(cmd[prevIndex:])
			} else if prevChar == "*" {
				result += decodeArrays(cmd[prevIndex:])
			}
		}
	}

	return result
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
		value := decode(msg)
		switch strings.ToUpper(value[0]) {
			case "PING":
				conn.Write([]byte(encodeSimpleStrings("PONG")))
				break
			case "ECHO":
				conn.Write([]byte(encodeSimpleStrings(value[1])))
				break
			case "GET":
				DataStoreMutex.Lock()
				dsVal, exists := DataStore[value[1]]

				if !exists {
					return "$-1\r\n";
				} else if (dsVal != nil && time.Now() - dsVal.ttl <= 0) {
					delete(DataStore, value[1])
					return "$-1\r\n";
				} else {
					conn.Write([]byte(encodeSimpleStrings(DataStore[value[1]])))
				}

				DataStoreMutex.Unlock()
				break
			case "SET":
				DataStoreMutex.Lock()
				DataStore[value[1]] = DataStoreValue{ Value: value[2] }
				if len(value > 3){
					expiry := strconv.Atoi(value[4])
					DataStore[value[1]].ttl = time.Now().Add(expiry * time.Millisecond)
				}
				DataStoreMutex.Unlock()
				conn.Write([]byte(encodeSimpleStrings("OK")))
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
