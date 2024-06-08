package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
}

var (
	clients    = make(map[net.Conn]Client)
	mutex      = sync.Mutex{}
	broadcast  = make(chan string) //A channel is a communication mechanism that allows you to pass data between goroutines (concurrent functions) safely.
	newClients = make(chan net.Conn)
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	client := Client{conn: conn}
	clients[conn] = client

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			delete(clients, conn)
			return
		}

		broadcast <- fmt.Sprintf("%s: %s", client.name, msg)
	}
}

func broadcastMessages() {
	for msg := range broadcast {
		for conn := range clients {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				delete(clients, conn)
			}
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	go broadcastMessages()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
