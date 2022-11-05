package problem6

import (
	"bufio"
	"fmt"
	"net"
	"protohackers/util"

	"github.com/google/uuid"
)

var (
	server   ServerWithReader
	observer *PlateObserver
	clients  []*Client

	pendingTickets Tickets
)

func init() {
	server = ServerWithReader{
		util.ServerTCP{
			ConnectionHandler: handle,
		},
	}
	observer = &PlateObserver{
		observer: make(chan Observation),
	}
}

func Problem() {
	go observer.listenForPlates()
	server.Start()
}

func handle(conn net.Conn) {

	clientId := uuid.New().String()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	client := &Client{
		id:   clientId,
		conn: conn,
		writer: serverWriter{
			w: writer,
		},
		reader: serverReader{
			r: reader,
		},
	}
	defer client.closeConnection()
	clients = append(clients, client)

	for {
		line, err := server.ReadMessage(client)
		if err != nil {
			if err == ErrUnknown {
				// throw error
				er := ErrorM{
					Message: fmt.Sprintf("Unknown protocol %s", err),
				}
				er.writeError(client.writer.w)
				return
			}
		}
		if line != nil {
			server.handleLine(line, client)
		}
	}
}

// Handling the read incoming messages
func (s *ServerWithReader) handleLine(line interface{}, client *Client) {
	switch line := line.(type) {
	case Plate:
		camera := client.camera
		observer.makeObservation(camera, line)

	case Dispatcher:
		line = Dispatcher(line)
		client.dispatcher = &line

		var filtered []*Ticket
		if len(pendingTickets.Tickets) > 0 {
			pendingTickets.mu.Lock()
			defer pendingTickets.mu.Unlock()
			for _, pending := range pendingTickets.Tickets {
				if client.dispatcher.hasRoad(pending.Road) {
					pending.writeTicket(client.writer.w)
				} else {
					filtered = append(filtered, pending)
				}
			}
			pendingTickets.Tickets = filtered
		}
	case Camera:
		line = Camera(line)
		client.camera = &line
	case Heartbeat:
		line = Heartbeat(line)
		if line.interval > 0 {
			client.heartbeat = &line
			client.scheduleHeartbeat()
		}
	default:
		fmt.Println("Help ", line)
	}
}

// Checking if a dispatcher has a road
func (d *Dispatcher) hasRoad(roadNmbr uint16) bool {
	for _, road := range d.Roads {
		if road == roadNmbr {
			return true
		}
	}
	return false
}
