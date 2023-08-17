package problem11

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

func writeByte(buffer *bytes.Buffer, b byte) {
	buffer.WriteByte(b)
}

func writeString(buffer *bytes.Buffer, s string) {
	writeUint32(buffer, uint32(len(s)))
	buffer.WriteString(s)
}

func writeUint16(buffer *bytes.Buffer, i uint16) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], i)
	buffer.Write(buf[:])
}

func writeUint32(buffer *bytes.Buffer, i uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], i)
	buffer.Write(buf[:])
}

func readString(r *bufio.Reader, length uint32) (string, error) {

	buff := make([]byte, length)
	_, err := io.ReadFull(r, buff)
	if err != nil {
		return "", err
	}
	return string(buff), nil
}

func readUint16(r *bufio.Reader) (uint16, error) {
	var n uint16
	err := binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return n, err
	}

	return n, nil
}

func readUint32(r *bufio.Reader) (uint32, error) {
	var n uint32
	err := binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return n, err
	}
	return n, nil
}
