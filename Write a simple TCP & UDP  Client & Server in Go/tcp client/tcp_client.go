package main

import (
	"fmt"
	"net"
)

func main() {
	//connect to server
	c, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	//send some data
	_, err = c.Write([]byte("Hello World"))
	if err != nil {
		fmt.Println(err)
		return
	}

	//close connection
	c.Close()
}
