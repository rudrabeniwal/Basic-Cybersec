package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
}

var (
	//map[keyType]valueType
	clients     = make(map[net.Conn]string)
	clientMutex sync.Mutex          //Type: sync.Mutex | Purpose: This variable is a mutex (short for mutual exclusion) used for synchronization. | Usage: Ensures that access to the clients map is safe from concurrent reads and writes by multiple goroutines. Without proper synchronization, accessing or modifying shared data concurrently could lead to race conditions or inconsistent states.
	broadcastCh = make(chan string) //Type: chan string | Purpose: This variable is a channel used for broadcasting messages to all connected clients. |  Usage: Allows the server to send messages to all clients simultaneously. Channels in Go are a powerful concurrency primitive, facilitating communication between goroutines. When a message is sent to this channel (broadcastCh <- message), it can be received by multiple clients simultaneously (through a separate goroutine listening on this channel).
)

func handleClient(client Client) {
	defer func() {

		/*

			clientMutex.Lock()
			Purpose: This acquires a lock on the clientMutex mutex.
			Functionality:
			When a goroutine calls clientMutex.Lock(), it attempts to gain exclusive access to the critical section protected by clientMutex.
			If another goroutine already holds the lock (clientMutex is locked), the current goroutine will block until it can acquire the lock.
			Once the lock is acquired, the goroutine can safely proceed to modify the clients map without interference from other goroutines.

			clientMutex.Unlock()
			Purpose: This releases the lock on the clientMutex mutex.
			Functionality:
			After the critical section (modifying the clients map) is complete, the goroutine calls clientMutex.Unlock() to release the lock.
			Releasing the lock allows other goroutines waiting for the lock to acquire it and proceed with their operations.
			It's crucial to unlock the mutex after finishing the critical section to avoid unnecessary blocking of other goroutines and to ensure concurrency.

		*/

		client.conn.Close()
		clientMutex.Lock()
		delete(clients, client.conn)
		clientMutex.Unlock()
		broadcast(fmt.Sprintf("%s has left the chat\n", client.name))
	}()

	client.conn.Write([]byte("Enter your name:\n"))
	name, _ := bufio.NewReader(client.conn).ReadString('\n')
	client.name = strings.TrimSpace(name) //strings.TrimSpace(name) removes any leading or trailing whitespace characters from the client's input (name).
	clientMutex.Lock()
	clients[client.conn] = client.name
	clientMutex.Unlock()

	broadcast(fmt.Sprintf("%s has joined the chat\n", client.name))

	for {
		message, err := bufio.NewReader(client.conn).ReadString('\n')
		if err != nil {
			return
		}
		broadcast(fmt.Sprintf("%s: %s", client.name, message))
	}
}

func broadcast(message string) {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	for conn := range clients {
		conn.Write([]byte(message))
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting TCP Server")
		os.Exit(1)

	}
	defer listener.Close()

	// Goroutine for broadcasting messages
	go func() {
		for {
			message := <-broadcastCh
			broadcast(message)
		}
	}()

	fmt.Println("Server Started on :8080")

	// Accepting and handling incoming client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		client := Client{conn: conn}
		go handleClient(client)
	}
}
