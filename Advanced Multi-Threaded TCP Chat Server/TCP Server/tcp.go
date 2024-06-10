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
			for _, client := range clients {
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

	conn.Write([]byte("Enter your password: "))
	password, _ := bufio.NewReader(conn).ReadString('\n')
	password = strings.TrimSpace(password)

	if auth, ok := authenticate(conn, username, password); !auth {
		return
	} else {
		authenticated[conn] = username
	}

	client := Client{conn: conn, name: username, room: ""}
	register <- client

	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			unregister <- conn
			break
		}
		handleClientMessage(client, strings.TrimSpace(msg))
	}
}

func authenticate(conn net.Conn, username, password string) (bool, error) {
	mu.Lock()
	defer  mu.Unlock()

	if storedPassword, ok := users[username]; ok{
		if storedPassword == password {
			conn.Write([]byte("Authentication Successful.\n"))
			return true, nil
		} else {
			conn.Write([]byte("Invalid Password.\n"))
			return false, nil
		}
	} else {
		users[username] = password
		conn.Write([]byte("Registration successful. You are now authenticated.\n"))
		return true, nil
	}
}

func handleClientMessage(client Client, msg string) {
	if strings.HasPrefix(msg, "/"){
		handleCommand(client, msg)
	} else {
		mu.Lock()
		if client.room != "" {
			broadcast <- fmt.Sprintf("[%s] %s: %s\n", client.room, client.name, msg)
		} else {
			client.conn.Write([]byte("Join a room to start chatting. Use /join <room> to join a room.\n"))
		}
		mu.Unlock()
	}
}

func handleCommand(client Client, msg string) {
	parts := strings.Split(msg, " ", 3)
	switch parts[0] {
	case "/list":
		client.conn.Write([]byte("Connected users: \n"))
		mu.Lock()
		for _,c := range clients {
			client.conn.Write([]byte(c.name + "\n"))
		}
		mu.Unlock()
	case "/msg":
		if len(parts) < 3 {
			client.conn.Write([]byte("Usage: /msg <user> <messege>\n"))
			return
		}
		targetName, messege := parts[1], parts[2]
		mu.Lock()
		for _, c := clients {
			if c.name == targetName {
				c.conn.Write([]byte(fmt.Sprintf("Private from %s: %s\n", client.name, message)))
				client.conn.Write([]byte(fmt.Sprintf("Private to %s: %s\n", targetName, message)))
				mu.Unlock()
				return
			}
		}
		mu.Unlock()
		client.conn.Write([]byte(fmt.Sprintf("User %s not found\n", targetName)))
	case "/join":
		if len(parts) < 2 {
			client.conn.Write([]byte("Usage: /join <room>\n"))
			return
		}
		room := parts[1]
		mu.Lock()
		if client.room != "" {
			oldRoomClients := rooms[client.room]
			for i, roomClient := range oldRoomClients {
				if roomClient.conn == client.conn {
					rooms[client.room] = append(oldRoomClients[:i], oldRoomClients[i+1:]... )
					break
				}
			}
		}
		client.room = room
		rooms[room] = append(rooms[room], client)
		clients[client.conn] = client
		client.conn.Write([]byte(fmt.Sprintf("Joined room %s\n", room)))
		mu.Unlock()
	default:
		client.conn.Write([]byte("Unknown command\n"))
	}
}