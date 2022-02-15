package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"
)

/////////////////////////////////////////////
// Simple Strings +Messaga\r\n
func encodeSimpleStrings(str string) string {
	return "+" + str + "\r\n"
}

func decodeSimpleStrings(cmd string) string {
	return cmd[1 : len(cmd)-6]
}

// Bulk Strings $4\r\nPing\r\n
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

// Arrays *3\r\n$3\r\nSET\r\n$4\r\nDING\r\n$4DONG\r\n
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

func decode(msg string) string {
	switch msg[0] {
	case '+':
		return decodeSimpleStrings(msg)
	case '$':
		return decodeBulkStrings(msg)
	case '*':
		return decodeArrays(msg)
	default:
		return ""
	}
}

func handleCients(conn net.Conn) string {
	var buffer [512]byte
	length, err := conn.Read(buffer[:])
	if err != nil {
		fmt.Println("Failed to read input" + err.Error())
	}
	msg := string(buffer[:length])

	return msg
}

func main() {
	fmt.Println("Your code goes here!")

	// Create a TCP Server
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	// Communicate
	for {
		// Accept the client connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		for {
			msg := handleCients(conn)

			switch msg[8 : len(msg)-2] {
			// *1\r\n$4\r\nping\r\n
			case "ping":
				conn.Write([]byte(encodeSimpleStrings("PONG")))
			}
		}
	}
}
