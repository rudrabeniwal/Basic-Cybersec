package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
	room string
}

var (
	clients = make(map[net.Conn]Client)
	rooms = make(map[string][]Client)
	broadcast = make(chan string)
	register = make(chan Client)
	unregister = make(chan net.Conn)
	mu sync.Mutex
	shutdown = make(chan bool)
	users = map[string]string{} //Username:Password
	authenticated = make(map[net.Conn]string)
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Server started on port 8080")

	go handleMesseges()

	go func() {
		http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			fmt.Fprintf(w,  "Connected clients: %d\n", len(clients))
			for _, client := range clients {
				fmt.Fprintf(w, "User: %s, Room: %s\n", client.name,client.room)
			}
		})
		log.Fatal(http.ListenAndServe(":8081", nil))
	} ()

		go func() {
			var input string
			for {
				fmt.Scanln(&input)
				if input == "exit" {
					shutdown <- true
					return
				}
			}
		} ()
	
		for {
			select {
			case <- shutdown:
				ln.Close()
				mu.Lock()
				for _, client := range clients {
					client.conn.Write([]byte("Server is shutting down...\n"))
					client.conn.Close()
				}
				mu.Unlock()
				return
			default:
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println("Error accepting connection:", err)
					continue
				}
				go handleConnection(conn)
			}
		}
}

func handleMesseges() {
	for {
		select {
		case msg := <- broadcast:
			mu.Lock()
			for _, client := clients {
				if client.room != "" {
					for _, roomClient := range rooms[client.room] {
						roomClient.conn.Write([]byte(msg))
					}
				}
			}
			mu.Unlock()

		case client := <- register:
			mu.Lock()
			clients[clinet.conn] = client
			if client.room != "" {
				rooms[client.room] = append(rooms[client.room], client)
			}
			mu.Unlock()
			broadcast <- fmt.Sprintf("%s joined the chat\n", client.name)
		
		case conn := <- unregister:
			mu.Lock()
			if client, ok := clients[conn]; ok {
				delete(client, conn)
				if client.room != "" {
					roomClients := rooms[client.room]
					for i, rommClient := range roomClients {
						if roomClient.conn == conn {
							romms[client.room] = append(roomClients[:i], roomClients[i+1:]...)
							break
						}
					}
				}
				broadcast <- fmt.Sprintf("%s left the chat\n", client.name)
				conn.Close()
			}
			mu.Unlock()
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Enter your username: "))
	username, _ := bufio.NewReader(conn).ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write()
}