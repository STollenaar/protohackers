package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
)

type isPrime struct {
	Method *string  `json:"method,"`
	Number *float64 `json:"number,"`
	Prime  bool     `json:"prime"`
}

func main() {
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
		go handle(conn)
	}
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
