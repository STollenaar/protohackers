package problem0

import (
	"fmt"
	"io"
	"net"
	"protohackers/util"
)

var server util.Server

func init() {
	server = util.Server{
		ConnectionHandler: handle,
	}
}

func Problem() {
	server.Start()
}

func handle(conn net.Conn) {
	defer conn.Close()

	if _, err := io.Copy(conn, conn); err != nil {
		fmt.Println("copy: ", err.Error())
	}
}
