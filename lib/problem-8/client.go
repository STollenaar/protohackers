package problem8

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
)

type serverReader struct {
	r   *bufio.Reader
	err error
}

type serverWriter struct {
	w *bufio.Writer
}

type Client struct {
	id         string
	cypherSpec []Cypher

	conn   net.Conn
	writer serverWriter
	reader serverReader
}

// Cleaning up all the connections
func (c *Client) closeConnection() {
	defer c.conn.Close()
}

func (e *serverReader) readString() string {
	if e.err != nil {
		return ""
	}

	n, err := e.r.ReadByte()
	if err != nil {
		e.err = err
		return ""
	}

	return e.byteToString(n)
}

func (e *serverReader) byteToString(n byte) string {
	buff := make([]byte, n)
	_, err := io.ReadFull(e.r, buff)
	if err != nil {
		e.err = err
		fmt.Println(err)
		return ""
	}
	return string(buff)
}

func (e *serverReader) readUint16() uint16 {
	if e.err != nil {
		return 0
	}
	var n uint16
	e.err = binary.Read(e.r, binary.BigEndian, &n)
	return n
}

func (e *serverReader) readUint32() uint32 {
	if e.err != nil {
		return 0
	}
	var n uint32
	e.err = binary.Read(e.r, binary.BigEndian, &n)
	return n
}

// func (t *Ticket) writeTicket(w *bufio.Writer) {
// 	writeByte(w, TicketType)
// 	writeString(w, t.Plate)
// 	writeUint16(w, t.Road)
// 	writeUint16(w, t.Mile1)
// 	writeUint32(w, t.Timestamp1)
// 	writeUint16(w, t.Mile2)
// 	writeUint32(w, t.Timestamp2)
// 	writeUint16(w, t.Speed)
// 	w.Flush()
// }

func writeByte(w *bufio.Writer, b byte) {
	w.WriteByte(b)
}

func writeString(w *bufio.Writer, s string) {
	if len(s) > math.MaxUint8 {
		s = s[:math.MaxUint8-3] + "..."
	}
	w.WriteByte(byte(len(s)))
	w.WriteString(s)
}

func writeUint16(w *bufio.Writer, i uint16) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], i)
	w.Write(buf[:])
}

func writeUint32(w *bufio.Writer, i uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], i)
	w.Write(buf[:])
}
