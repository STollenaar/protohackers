package problem10

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type Client struct {
	reader *bufio.Reader
	writer net.Conn
}

func (c *Client) closeConnection() {
	defer c.writer.Close()
}

func (c *Client) send(line string, args ...any) {
	if !strings.HasSuffix(line, "\n") {
		line = line + "\n"
	}

	io.WriteString(c.writer, fmt.Sprintf(line, args...))
}

func (c *Client) sendRaw(data []byte) {
	c.writer.Write(data)
}

func (c *Client) readLine() Request {
	line, err := c.reader.ReadString('\n')
	fields := strings.Fields(line)
	if err != nil || len(fields) == 0 {
		return Request{}
	}
	return Request{strings.ToUpper(fields[0]), fields[1:]}
}

func (c *Client) readLength(n int) []byte {
	data := make([]byte, n)
	io.ReadFull(c.reader, data)
	return data
}
