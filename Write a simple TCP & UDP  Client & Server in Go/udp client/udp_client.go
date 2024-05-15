package main

import (
	"fmt"
	"net"
)

func main() {
	//resolve udp address
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	//create udp connection
	//func net.DialUDP(network string, laddr *net.UDPAddr, raddr *net.UDPAddr) (*net.UDPConn, error)
	//DialUDP acts like Dial for UDP networks.
	//The network must be a UDP network name; see func Dial for details.
	//If laddr is nil, a local address is automatically chosen. If the IP field of raddr is nil or an unspecified IP address, the local system is assumed.

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	//send messege to server
	messege := []byte("Hello UDP Server!")
	_, err = conn.Write(messege)
	if err != nil {
		fmt.Println(err)
		return
	}

	//receive response from server
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("received from server:", string(buffer[:n]))
}
