package problem2

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"protohackers/util"
)

type RequestType byte
type Request struct {
	Type      RequestType
	TimeStamp int32
	Price     int32
}

type ServerWithReader struct {
	util.Server
}

const (
	RequestInsert RequestType = 'I'
	RequestQuery  RequestType = 'Q'
)

var server ServerWithReader

func (s *ServerWithReader) readMessage(conn net.Conn) (incoming Request, err error) {

	err = binary.Read(conn, binary.BigEndian, &incoming)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
			return Request{}, err
		}
	}
	return incoming, err
}

func init() {
	server = ServerWithReader{
		Server: util.Server{
			ConnectionHandler: handle,
		},
	}
}

func Problem() {
	server.Start()
}

func handle(conn net.Conn) {
	defer conn.Close()

	assets := make(map[int32]int32)
	for {
		message, err := server.readMessage(conn)
		if err != nil {
			return
		}
		fmt.Println(message)

		switch message.Type {
		case RequestInsert:
			assets[message.TimeStamp] = message.Price
		case RequestQuery:
			if message.Price < message.TimeStamp {
				binary.Write(conn, binary.BigEndian, int32(0))
				continue
			}
			mean := 0
			stamps := 0
			for k, v := range assets {
				if message.TimeStamp <= k && k <= message.Price {
					mean += int(v)
					stamps++
				}
			}

			if stamps == 0 {
				binary.Write(conn, binary.BigEndian, int32(0))
			} else {
				binary.Write(conn, binary.BigEndian, int32(mean/stamps))
			}
		default:
			log.Printf("Invalid request: %v", message.Type)
			conn.Write([]byte{})
			return
		}
	}
}
