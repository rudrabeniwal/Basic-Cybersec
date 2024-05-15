package main

import (
	"fmt"
	"net"
)

func main() {
	//listen on incoming connections on port 8080
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	//accept incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		//handle connection in new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	//close the connection when we're done
	defer conn.Close()
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	//print data
	fmt.Printf("%s", buffer)
}
