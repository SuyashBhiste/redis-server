package main

import (
	"bufio"
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

/////////////////////////////////////////////
func decode(cmd string) string {
	var decodedMsg string

	switch cmd[0] {
	case '+':
		decodedMsg = decodeSimpleStrings(cmd)
		break

	case '$':
		decodedMsg = decodeBulkStrings(cmd)
		break

	case '*':
		decodedMsg = decodeArrays(cmd)
		break
	}

	return decodedMsg
}

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

	// Communicate
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input")
		}
		fmt.Println("Input Msg 1: " + msg + strconv.Itoa(len(msg)))

		if msg[0] != '*' {
			fmt.Println("Ooops!!! It was not *")
			break
		}

		msg, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input")
		}
		fmt.Println("Input Msg 2: " + msg + strconv.Itoa(len(msg)))

		if msg[0] != '$' {
			fmt.Println("Ooops!!! It was not $")
			break
		}

		msg, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input")
		}
		fmt.Println("Input Msg 3: " + msg + strconv.Itoa(len(msg)))

		switch msg[0 : len(msg)-2] {
		case "PING":
			conn.Write([]byte(encodeSimpleStrings("PONG")))
			break
		}
	}
}
