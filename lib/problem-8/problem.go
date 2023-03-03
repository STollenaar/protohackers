package problem8

import (
	"bufio"
	"fmt"
	"net"
	"protohackers/util"
)

var (
	server ServerWithReader
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
	defer conn.Close()

	reader := bufio.NewReader(conn)
	encoder, decoder, err := server.ReadCypherSpec(reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	decodeReader := &decodeReader{r: reader, cypherSpec: decoder}
	decodeReader.resetCypherspec()

	readScanner := bufio.NewScanner(decodeReader)

	writer := encodeWriter{w: conn, cypherSpec: encoder}
	writer.resetCypherspec()

	for readScanner.Scan() {
		if readScanner.Err() != nil {
			return
		}

		line := readScanner.Text()

		toys := createToys(line)
		max := findMaxToy(toys)

		msg := []byte(max.toString())
		writer.Write(msg)
	}
}
