package main

import (
	"errors"
	"fmt"
	"net"
	"os"
)

var EOF = errors.New("EOF")

func main() {
	ln, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		fmt.Println("listen: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("listening on port " + os.Getenv("PORT"))
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("connection from ", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 32*1024)
	for {
		messageLength, err := conn.Read(buff)

		if err != nil {
			if err != EOF {
				fmt.Println("Error reading data: ", err)
			}
			break
		}

		_, err = conn.Write(buff[0:messageLength])

		if err != nil {
			fmt.Println("Error writing data: ", err)
			break
		}
	}
}
