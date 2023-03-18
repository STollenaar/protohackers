package problem10

import (
	"bufio"
	"net"
	"protohackers/util"
)

var (
	server ServerWithReader
	files  *File
)

func init() {
	server = ServerWithReader{
		util.ServerTCP{
			ConnectionHandler: handle,
		},
	}
	files = newFiles()
}

func Problem() {
	server.Start()
}

func handle(conn net.Conn) {

	client := &Client{
		reader: bufio.NewReader(conn),
		writer: conn,
	}
	defer client.closeConnection()

	for {
		client.send("READY")

		req := client.readLine()
		server.handleLine(req, client)
	}
}
