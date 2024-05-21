package main

import (
	"fmt"
	"net"
)

func main() {
	// resolve UDP address
	//ResolveUDPAddr returns an address of UDP end point.
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	// create UDP listener
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	buffer := make([]byte, 1024)
	//handle incoming messeges
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Data received : %s", string(buffer[:n]))

		//echo back to client
		messege := []byte("Hello UDP Client!!")
		_, err = conn.WriteToUDP(messege, addr)
		if err != nil {
			fmt.Println(err)
			continue
		}

	}

}
