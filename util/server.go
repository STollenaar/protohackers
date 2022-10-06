package util

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

type ServerTCP struct {
	ConnectionHandler func(net.Conn)
}

func (s *ServerTCP) Start() {
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
		go s.ConnectionHandler(conn)
	}
}

type ServerUDP struct {
	ConnectionHandler func(*net.UDPConn)
}

func (s *ServerUDP) Start() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("listening on port " + os.Getenv("PORT"))
	s.ConnectionHandler(conn)
}
