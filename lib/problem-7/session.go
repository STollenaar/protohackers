package problem7

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SendReceipt struct {
	totalValue int64 // Also the POS in the data string
	nextTotal  int64
	length     int
	data       string
	retransmit *time.Ticker
}

type Message struct {
	length int // Message length
	pos    int
	data   string // The data string of the incoming message
}

type Messages struct {
	mu     sync.Mutex
	buffer map[int64]*Message // Slice of incoming data from client to server. Need to be referenced by incoming data messages
}

type Session struct {
	sessionId int64
	timeout   *time.Timer

	totalSendSize int64
	sendReceipts  []*SendReceipt // Slice of outgoing data from server to client. Need to be referenced by incoming ack messages

	totalBufferSize int
	messageBuffer   *Messages
}

func (s *Session) addToBuffer(message *Message, pos int64) {
	s.messageBuffer.mu.Lock()
	defer s.messageBuffer.mu.Unlock()
	s.messageBuffer.buffer[pos] = message
	s.totalBufferSize += message.length
}

func (s *Session) createSend(data string, pos int64, conn *net.UDPConn, remote *net.UDPAddr) int64 {

	ps := strings.Split(data, "\n")
	totLen := totalLength(ps) + strings.Count(data, "\n")
	data = strings.ReplaceAll(data, `\`, `\\`)
	data = strings.ReplaceAll(data, `/`, `\/`)
	sendData := fmt.Sprintf("/data/%d/%d/%s/", s.sessionId, pos, data)

	receipt := SendReceipt{
		totalValue: pos,
		nextTotal:  pos + int64(totLen),
		length:     totLen,
		data:       sendData,
	}
	receipt.scheduleRetransmit(s, conn, remote)
	s.sendReceipts = append(s.sendReceipts, &receipt)
	return int64(totLen)
}

func (s *Session) send(data string, conn *net.UDPConn, remote *net.UDPAddr) {
	fmt.Printf("<Sending SessionID: %d> Size: %d, Data: %s\n", s.sessionId, len([]byte(strconv.Quote(data))), strconv.Quote(data))
	_, err := conn.WriteTo([]byte(data), remote)
	if err != nil {
		panic(err)
	}
}

func (s *Session) scheduleTimeout() {
	ticker := time.NewTimer(time.Duration(time.Second * 60))
	if s == nil {
		return
	}
	if s.timeout != nil {
		s.timeout.Stop()
	}
	s.timeout = ticker
	go func() {
		for {
			<-ticker.C
			// do stuff
			for _, sr := range s.sendReceipts {
				sr.retransmit.Stop()
			}
		}
	}()
}

func (s *Session) close() {
	if s == nil {
		return
	}
	if s.timeout != nil {
		s.timeout.Stop()
	}
	for _, sr := range s.sendReceipts {
		sr.retransmit.Stop()
	}
}

func (s *SendReceipt) scheduleRetransmit(session *Session, conn *net.UDPConn, remote *net.UDPAddr) {
	ticker := time.NewTicker(time.Duration(time.Second * 2))

	if s.retransmit != nil {
		s.retransmit.Stop()
	}
	s.retransmit = ticker
	session.send(s.data, conn, remote)

	go func() {
		for {
			<-ticker.C
			session.send(s.data, conn, remote)
		}
	}()
}

func (s *Session) getLargestMessage() (int64, *Message) {
	s.messageBuffer.mu.Lock()
	defer s.messageBuffer.mu.Unlock()
	if len(s.messageBuffer.buffer) == 0 {
		return 0, nil
	}

	max := int64(0)
	for pos := range s.messageBuffer.buffer {
		if pos > max {
			max = pos
		}
	}
	return max, s.messageBuffer.buffer[max]
}

func (s *Session) getLargestSendReceipt() *SendReceipt {
	if len(s.sendReceipts) == 0 {
		return &SendReceipt{}
	}
	max := s.sendReceipts[0]

	for _, r := range s.sendReceipts {
		if r.totalValue > max.totalValue {
			max = r
		}
	}
	return max
}

func (s *Session) getFullBuffer() (message string) {
	var keys []int

	for i := range s.messageBuffer.buffer {
		keys = append(keys, int(i))
	}
	sort.Ints(keys)

	for _, i := range keys {
		m := s.messageBuffer.buffer[int64(i)]
		fmt.Printf("<Building message %d>Index: %d\n", s.sessionId, i)
		message += m.data
	}
	return message
}
