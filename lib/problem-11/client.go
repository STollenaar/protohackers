package problem11

import (
	"bufio"
	"net"
)

type serverReader struct {
	r *bufio.Reader
}

type serverWriter struct {
	w *bufio.Writer
}

type Client struct {
	conn   net.Conn
	writer serverWriter
	reader serverReader
}
