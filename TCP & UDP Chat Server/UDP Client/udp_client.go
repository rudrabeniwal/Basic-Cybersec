package main

//os.Exit(1) terminates the program immediately. Unlike return, which unwinds the stack and allows deferred functions to run, os.Exit does not execute any deferred functions. This is crucial when you need to ensure that the program stops right away without executing further code.

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func listenForMessages(conn *net.UDPConn) {
	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading form UDP Server:", err)
			return
		}
		fmt.Print(string(buf[:n]) + "\n")
	}
}

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Conncected UDP server on 8080")

	go listenForMessages(conn)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/quit" {
			break
		}

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}

	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from stdin:", err)
	}
}
