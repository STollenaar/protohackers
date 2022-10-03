package util

import (
	"fmt"
	"net"
	"os"
)

type Server struct {
	ConnectionHandler func(net.Conn)
}

func (s *Server) Start() {
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
