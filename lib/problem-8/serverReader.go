package problem8

import (
	"bufio"
	"bytes"
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
	ErrInvalid = errors.New("invalid cypherspec provided")
)

type ServerWithReader struct {
	util.ServerTCP
}

func (s *ServerWithReader) ReadCypherSpec(reader *bufio.Reader) (encoder, decoder []Cypher, err error) {
	var rawCyphers []byte

	for _, err := reader.Peek(1); err == nil; {
		r, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
			return nil, nil, err
		}

		if r == EndType {
			if len(rawCyphers) > 0 && rawCyphers[len(rawCyphers)-1] != AddNType && rawCyphers[len(rawCyphers)-1] != XORNType {
				break
			} else if len(rawCyphers) == 0 {
				break
			}
		}
		rawCyphers = append(rawCyphers, r)
	}

	for i := 0; i < len(rawCyphers); i++ {
		switch rawCyphers[i] {
		case ReverseType:
			encoder = append(encoder, &ReverseCypher{})
			decoder = append(decoder, &ReverseCypher{})
		case XORNType:
			i++
			value := rawCyphers[i]
			encoder = append(encoder, &XORNCypher{value: value})
			decoder = append(decoder, &XORNCypher{value: value})
		case XORPosType:
			encoder = append(encoder, &XORPosCypher{pos: 0})
			decoder = append(decoder, &XORPosCypher{pos: 0})
		case AddNType:
			i++
			value := rawCyphers[i]
			encoder = append(encoder, &AddNCypher{value: value})
			decoder = append(decoder, &AddNCypher{value: value})
		case AddPosType:
			encoder = append(encoder, &AddPosCypher{pos: 0})
			decoder = append(decoder, &AddPosCypher{pos: 0})
		}
	}
	if s.isNoOp(encoder) {
		return nil, nil, ErrInvalid
	}
	return encoder, decoder, nil
}

func (s *ServerWithReader) isNoOp(cypherSpec []Cypher) bool {
	test := []byte("hello")
	var encoded []byte
	for _, i := range test {
		for _, cyp := range cypherSpec {
			i = cyp.operation(i, true)
		}
		encoded = append(encoded, i)
	}
	return bytes.Equal(encoded, test)
}
