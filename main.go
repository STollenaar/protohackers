package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
)

var EOF = errors.New("EOF")

type isPrime struct {
	Method string  `json:"method,omitempty"`
	Number float64 `json:"number,omitempty"`
	Prime  bool    `json:"prime"`
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

	buff := make([]byte, 32*1024)
	for {
		rl, err := conn.Read(buff)

		if err != nil {
			if err != EOF {
				fmt.Println("Error reading data: ", err)
			}
			break
		}

		data := new(isPrime)

		fmt.Println("Found string: ", string(buff[:rl]))
		err = json.Unmarshal(buff[:rl], data)
		if err != nil {
			conn.Write([]byte(err.Error()))
			break
		}

		if float64(int64(data.Number)) != data.Number {
			conn.Write([]byte("Number was not an integer"))
			break
		}
		if data.Method != "isPrime" {
			conn.Write([]byte("method was not isPrime"))
			break
		}
		if big.NewInt(int64(data.Number)).ProbablyPrime(0) {
			data.Prime = true
		} else {
			data.Prime = false
		}
		out, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error writing data: ", err)
			break
		}
		_, err = conn.Write(out)

		if err != nil {
			fmt.Println("Error writing data: ", err)
			break
		}
	}
}
