package problem9

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"protohackers/util"

	"github.com/google/uuid"
)

var (
	server ServerWithReader
	queues QueueMap
)

func init() {
	server = ServerWithReader{
		util.ServerTCP{
			ConnectionHandler: handle,
		},
	}
	queues = *NewQueueMap()
}

func Problem() {
	server.Start()
}

func handle(conn net.Conn) {
	reader := bufio.NewReader(conn)

	buffer := bufio.NewScanner(reader)

	client := &Client{
		conn: conn,
		id:   uuid.NewString(),
	}
	defer client.closeConnection()

	for buffer.Scan() {

		line := buffer.Text()

		request := new(Request)
		err := json.Unmarshal([]byte(line), request)
		fmt.Printf("[Request %s] %v\n", client.id, request)

		if err != nil {
			response := Response{
				Status: "error",
				Error:  err.Error(),
			}
			resp, _ := json.Marshal(response)
			fmt.Printf("[Response %s] %v\n", client.id, response)
			conn.Write(append(resp, []byte("\n")...))
			continue
		}

		response := server.handleLine(request, client)
		resp, _ := json.Marshal(response)
		fmt.Printf("[Response %s] %v\n", client.id, response)
		conn.Write(append(resp, []byte("\n")...))
	}
}
