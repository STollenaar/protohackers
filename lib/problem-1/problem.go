package problem1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"protohackers/util"
)

var server util.Server

type isPrime struct {
	Method *string  `json:"method,"`
	Number *float64 `json:"number,"`
	Prime  bool     `json:"prime"`
}

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

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		buff := scanner.Bytes()

		if os.Getenv("DEBUG") == "true" {
			fmt.Println("Read string: ", string(buff))
		}

		data := new(isPrime)

		err := json.Unmarshal([]byte(buff), data)
		if err != nil {
			if os.Getenv("DEBUG") == "true" {
				fmt.Println("Malformed request: ", string(buff), " ", err)
			}
			conn.Write([]byte(err.Error()))
			return
		}

		if data.Method == nil || (data.Method != nil && *data.Method != "isPrime") || data.Number == nil {
			if os.Getenv("DEBUG") == "true" {
				fmt.Println("Malformed request: ", string(buff))
			}
			conn.Write([]byte("Malformed request\n"))
			return
		}

		if *data.Number != float64(int64(*data.Number)) {
			data.Prime = false
		} else if big.NewInt(int64(*data.Number)).ProbablyPrime(0) {
			data.Prime = true
		} else {
			data.Prime = false
		}

		out, err := json.Marshal(data)
		out = append(out, []byte("\n")...)

		if os.Getenv("DEBUG") == "true" {
			fmt.Println("Responding with: ", string(out))
		}

		if err != nil {
			fmt.Println("Error writing data: ", err)
			return
		}
		conn.Write(out)
	}
}
