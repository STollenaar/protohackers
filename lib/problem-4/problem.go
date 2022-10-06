package problem4

import (
	"fmt"
	"net"
	"protohackers/util"
	"strings"
)

var (
	server   util.ServerUDP
	database map[string]string
)

func init() {
	server = util.ServerUDP{
		ConnectionHandler: handle,
	}
	database = make(map[string]string)
}

func Problem() {
	server.Start()
}

func handle(conn *net.UDPConn) {
	defer conn.Close()

	for {
		message := make([]byte, 256)
		rlen, remote, err := conn.ReadFromUDP(message)
		if err != nil {
			panic(err)
		}
		data := string(message[:rlen])

		if data == "version" {
			conn.WriteTo([]byte("version=spices\n"), remote)
		} else if strings.Contains(data, "=") {
			slice := strings.Split(data, "=")
			key := slice[0]
			value := strings.Join(slice[1:], "=")
			database[key] = value
		} else {
			conn.WriteTo([]byte(fmt.Sprintf("%s=%s", data, database[data])), remote)
		}
		fmt.Printf("received: %s from %s\n", data, remote)
	}
}
