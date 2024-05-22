package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	clients     = make(map[string]*net.UDPAddr)
	clientMutex sync.Mutex
	broadcastCh = make(chan string)
)

func handleMesseges(conn *net.UDPConn) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			continue
		}

		clientMutex.Lock()
		if _, exists := clients[addr.String()]; !exists {
			clients[addr.String()] = addr
			broadcastCh <- fmt.Sprintf("New client joined: %s", addr)
		}
		clientMutex.Unlock()

		messege := string(buf[:n])
		broadcastCh <- fmt.Sprintf("%s: %s", addr, messege)
	}
}

func broadcastMessages(conn *net.UDPConn) {
	for message := range broadcastCh {
		clientMutex.Lock()
		for _, addr := range clients {
			conn.WriteToUDP([]byte(message), addr)
		}
		clientMutex.Unlock()
	}
}

func main() {
	addr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Eror starting the UDP server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("UDP Server started on :8080")

	go broadcastMessages(conn)
	handleMesseges(conn)
}
