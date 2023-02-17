package problem8

import (
	"errors"
	"fmt"
	"protohackers/util"
)

const (
	EndType     byte = 0x00
	ReverseType byte = 0x01
	XORNType    byte = 0x02
	XORPosType  byte = 0x03
	AddNType    byte = 0x04
	AddPosType  byte = 0x05
	ErrType     byte = 0x06
)

type ErrorM struct {
	Message string
}

var (
	ErrUnknown = errors.New("unknown message")
)

type ServerWithReader struct {
	util.ServerTCP
}

// Reading an incoming message
func (s *ServerWithReader) ReadMessage(client *Client) (msg string, err error) {
	for _, err = client.reader.r.Peek(1); err == nil; {
		typ, err := client.reader.r.ReadByte()
		fmt.Println("Type", typ, "ClientCypher", client.cypherSpec)
		if err != nil {
			fmt.Println("Error1", err)
			return "", err
		}

		switch typ {
		case ReverseType:
			fmt.Println("Reverse")
			client.cypherSpec = append(client.cypherSpec, ReverseCypher{})
		case XORNType:
			value, err := client.reader.r.ReadByte()
			fmt.Println("XORN with value", value, "Err", err)
			client.cypherSpec = append(client.cypherSpec, XORNCypher{value: value})
		case XORPosType:
			fmt.Println("XORPos")
			client.cypherSpec = append(client.cypherSpec, XORPosCypher{})
		case AddNType:
			value, err := client.reader.r.ReadByte()
			fmt.Println("AddN with value", value, "Err", err)
			client.cypherSpec = append(client.cypherSpec, AddNCypher{value: value})
		case AddPosType:
			fmt.Println("AddPos")
			client.cypherSpec = append(client.cypherSpec, XORPosCypher{})
		case EndType:
		default:
			db := s.decodeByte(client, typ)
			fmt.Println("Decoded", db, "Message:", string(db))
			if string(db) == "\n" {
				fmt.Println("Done decoding, full message", msg)
				return msg, nil
			}
			msg += string(db)
		}
	}
	return msg, nil
}

func (s *ServerWithReader) decodeByte(client *Client, b byte) (msg byte) {
	for _, cypher := range client.cypherSpec {
		b = cypher.operation(b, 0)
	}
	return b
}
