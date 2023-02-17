package problem8

import (
	"bufio"
	"fmt"
	"net"
	"protohackers/util"

	"github.com/google/uuid"
)

var (
	server  ServerWithReader
	clients []*Client
)

func init() {
	server = ServerWithReader{
		util.ServerTCP{
			ConnectionHandler: handle,
		},
	}
}

func Problem() {
	server.Start()
}

func handle(conn net.Conn) {

	clientId := uuid.New().String()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	client := &Client{
		id:   clientId,
		conn: conn,
		writer: serverWriter{
			w: writer,
		},
		reader: serverReader{
			r: reader,
		},
	}
	defer client.closeConnection()
	clients = append(clients, client)

	for {
		line, err := server.ReadMessage(client)
		fmt.Println(line, err)
		if err != nil {
			if err == ErrUnknown {
				client.closeConnection()
				return
			}
		}
		fmt.Println(line)
	}
}
