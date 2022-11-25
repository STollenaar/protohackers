package problem7

import (
	"fmt"
	"log"
	"net"
	"protohackers/util"
	"strconv"
	"strings"
	"sync"
)

var (
	server   util.ServerUDP
	sessions Sessions
)

type Sessions struct {
	mu       sync.Mutex
	sessions map[int64]*Session
}

func init() {
	server = util.ServerUDP{
		ConnectionHandler: handle,
	}
	sessions = Sessions{
		sessions: make(map[int64]*Session),
	}
}

func splitFunc(c rune) bool {
	return c == '/'
}

func Problem() {
	server.Start()
	fmt.Println("I FOR SOME REASON ENDED")
}

func handle(conn *net.UDPConn) {
	defer conn.Close()

	for {
		message := make([]byte, 2000)
		rlen, remote, err := conn.ReadFromUDP(message)
		if err != nil {
			log.Fatal(err)
		}
		data := string(message[:rlen])
		if len(data) == 0 || string(data[0]) != "/" || string(data[len(data)-1]) != "/" {
			fmt.Printf("<Bad packet> %s\n", data)
			continue
		}
		fields := strings.FieldsFunc(data, splitFunc)
		if len(fields) < 2 {
			fmt.Printf("<Packet too small> %s\n", data)
			continue
		}
		sessionId, err := strconv.ParseInt(fields[1], 10, 32)
		if err != nil || sessionId > 2147483648 {
			fmt.Printf("<Bad session data> %v\n", err)
			continue
		}
		fmt.Printf("<Received SessionID: %d>: %s\n", sessionId, strconv.Quote(data))

		switch fields[0] {
		default:
			continue
		case "close":
			sessions.mu.Lock()
			if ses, ok := sessions.sessions[sessionId]; ok && ses != nil {
				ses.send(fmt.Sprintf("/close/%d/", sessionId), conn, remote)
				ses.close()
			}
			sessions.sessions[sessionId] = nil
			sessions.mu.Unlock()
		case "connect":
			sessions.mu.Lock()
			if _, ok := sessions.sessions[sessionId]; !ok {
				session := &Session{
					sessionId: sessionId,
					messageBuffer: &Messages{
						buffer: make(map[int64]*Message),
					},
				}
				sessions.sessions[sessionId] = session

			}
			sessions.sessions[sessionId].scheduleTimeout()
			sessions.sessions[sessionId].send(fmt.Sprintf("/ack/%d/0/", sessionId), conn, remote)
			sessions.mu.Unlock()
		case "ack":
			sessions.mu.Lock()
			session, ok := sessions.sessions[sessionId]
			if !ok {
				fmt.Printf("<session not ok %d>\n", sessionId)
				noSessionSend(fmt.Sprintf("/close/%d/", sessionId), sessionId, conn, remote)
				sessions.mu.Unlock()
				continue
			}
			pos, err := strconv.ParseInt(fields[2], 10, 32)
			session.scheduleTimeout()
			if err != nil || pos < 0 {
				sessions.mu.Unlock()
				continue
			}

			largestMessage := session.getLargestSendReceipt()
			if pos > session.totalSendSize {
				fmt.Printf("<POS above totalSendSize %d>POS: %d, totalSendSize %d\n", sessionId, pos, session.totalSendSize)

				session.send(fmt.Sprintf("/close/%d/", sessionId), conn, remote)
				session.close()
				sessions.sessions[sessionId] = nil
				sessions.mu.Unlock()
				continue
			}

			// POS is smaller than totalMessageValue
			// Skipping retransmit and relying on the auto retransmits
			if pos < largestMessage.nextTotal {
				fmt.Printf("<POS below lm next %d>LargestMessage: %d, POS: %d, totalSendSize %d\n", sessionId, largestMessage.nextTotal, pos, session.totalSendSize)

				// for _, receipt := range session.sendReceipts {
				// 	if receipt.totalValue > int(pos) {
				// 		session.send(receipt.data, conn, remote)
				// 	}
				// }
				sessions.mu.Unlock()
				continue
			}

			// Stopping sending of messages that are smaller than the POS
			for _, sr := range session.sendReceipts {
				if sr.nextTotal == pos {
					sr.retransmit.Stop()
				}
			}
			sessions.mu.Unlock()
		case "data":
			if len(fields) < 3 {
				fmt.Printf("<Ignoring bad data packet %d>\n", sessionId)
				continue
			}
			pos, err := strconv.ParseInt(fields[2], 10, 32)
			if err != nil || pos < 0 {
				fmt.Printf("<Ignoring bad pos %d>\n", sessionId)
				continue
			}
			sessions.mu.Lock()
			session, ok := sessions.sessions[sessionId]
			if !ok {
				fmt.Printf("<Ignoring bad session %d>\n", sessionId)
				noSessionSend(fmt.Sprintf("/close/%d/", sessionId), sessionId, conn, remote)
				sessions.mu.Unlock()
				continue
			}
			session.scheduleTimeout()

			dataString := strings.Join(fields[3:], `/`)

			if len(dataString) == 0 {
				fmt.Printf("<ACK empty %d> value %d\n", sessionId, session.totalBufferSize)
				session.send(fmt.Sprintf("/ack/%d/%d/", sessionId, session.totalBufferSize), conn, remote)
				sessions.mu.Unlock()
				continue
			}

			var ds string
			escaped := 0
			for err == nil && escaped < 5 {
				ds, err = strconv.Unquote(fmt.Sprintf("\"%s\"", dataString))
				if err == nil {
					dataString = ds
				}
				escaped++
			}
			for strings.Contains(dataString, `\\`) {
				dataString = strings.ReplaceAll(dataString, `\\`, `\`)
			}

			var terminating bool

			mod := strings.Count(dataString, "\n")
			if dataString[len(dataString)-1:] == "\n" {
				dataString = dataString[:len(dataString)-1]
				terminating = true
			} else if len(dataString) < 900 {
				sessions.mu.Unlock()
				continue
			}

			if session == nil {
				fmt.Println("SOMEHOW HERE")
				noSessionSend(fmt.Sprintf("/close/%d/", sessionId), sessionId, conn, remote)
				sessions.mu.Unlock()
				continue
			}
			dataString = strings.ReplaceAll(dataString, `\/`, "/")
			sentences := strings.Split(dataString, "\n")
			totLen := totalLength(sentences) + mod

			message := Message{
				length: totLen,
				data:   dataString,
			}

			if int(pos) == session.totalBufferSize+int(session.totalSendSize) {
				session.addToBuffer(&message, pos)
			} else if int(pos) < session.totalBufferSize+int(session.totalSendSize) {
				_, lm := session.getLargestMessage()
				if lm == nil {
					fmt.Printf("<Possible bad state. NIL message %d>POS: %d, totalBufferSize %d, totalSendSize: %d\n", sessionId, pos, session.totalBufferSize, session.totalSendSize)
					sessions.mu.Unlock()
					continue
				}
				if lm.length < totLen {
					session.totalBufferSize -= lm.length
					session.addToBuffer(&message, pos)
				} else {
					// Possinle outdated packet
					fmt.Printf("<Possible outdated packet %d>POS: %d\n", sessionId, pos)
					sessions.mu.Unlock()
					continue
				}
			} else if pos > session.totalSendSize {
				// Got more than I already got. Need retransmission
				lp, _ := session.getLargestMessage()
				fmt.Printf("<Data POS too big %d>POS: %d, totalBufferSize: %d, totalSendSize: %d, value %d\n", sessionId, pos, session.totalBufferSize, session.totalSendSize, lp)
				session.send(fmt.Sprintf("/ack/%d/%d/", sessionId, lp), conn, remote)
				sessions.mu.Unlock()
				continue
			}

			fmt.Printf("<ACK initial %d> value %d\n", sessionId, session.totalBufferSize+int(session.totalSendSize))
			session.send(fmt.Sprintf("/ack/%d/%d/", sessionId, session.totalBufferSize), conn, remote)

			if !terminating {
				// Current message wasn't terminating so stopping further execution
				sessions.mu.Unlock()
				continue
			}
			// Preparing to send data back
			session.messageBuffer.mu.Lock()

			fullMessage := session.getFullBuffer()
			fmt.Printf("<Datastring SessionID: %d>: %s\n", sessionId, strconv.Quote(fullMessage))
			sentences = strings.Split(fullMessage, "\n")
			reversedSent := reverseSlice(sentences)
			reversed := strings.Join(reversedSent, "\n")

			fmt.Printf("<Reversed Datastring SessionID: %d>: %s\n", sessionId, strconv.Quote(reversed))

			packets := chunks(reversed, 900)

			posSend := session.totalSendSize
			for i, packet := range packets {

				if i == len(packets)-1 && string(packet[len(packet)-1]) != "\n" {
					packet += "\n"
				}

				totLen := session.createSend(packet, posSend, conn, remote)
				posSend += totLen
			}
			session.messageBuffer.buffer = make(map[int64]*Message)
			session.totalBufferSize = 0
			session.totalSendSize = posSend
			session.messageBuffer.mu.Unlock()
			sessions.mu.Unlock()
		}
	}
}

func reverseSlice(slice []string) (reversed []string) {
	for _, s := range slice {
		reversed = append(reversed, reverse(s))
	}
	return reversed
}

func filter(slice []string) (filtered []string) {
	for _, s := range slice {
		if s != "\n" && s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func totalLength(slice []string) (length int) {
	filterd := filter(slice)
	for _, s := range filterd {
		length += len(s)
	}
	return length
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func chunks(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

func noSessionSend(data string, sessionId int64, conn *net.UDPConn, remote *net.UDPAddr) {
	fmt.Printf("<Sending SessionID: %d> Size: %d, Data: %s\n", sessionId, len([]byte(strconv.Quote(data))), strconv.Quote(data))
	_, err := conn.WriteTo([]byte(data), remote)
	if err != nil {
		panic(err)
	}
}
