package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	//connect to server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server ", err)
	}
	defer conn.Close()

	//Read messeges from server
	go func() {
		reader := bufio.NewReader(conn)
		for {
			messege, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed by server ", err)
			}
			fmt.Print(messege)
		}
	}()

	//read input from stdin and send it to server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/quit" {
			break
		}
		_, err := fmt.Fprintf(conn, "%s\n", text)
		/*
			How fmt.Fprintf Works
			Functionality:
			fmt.Fprintf formats and writes data to a specified io.Writer interface (conn in this case).
			The format specifier ("%s\n") indicates that text (a string variable containing the client's message) should be formatted as a string followed by a newline ('\n') character.
			This formatted string is then written to the server through the established TCP connection (conn).
		*/
		if err != nil {
			fmt.Println("Error sending messegw: ", err)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from stdin:", err)
	}
}
