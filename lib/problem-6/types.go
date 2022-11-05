package problem6

import (
	"sync"
	"time"
)

type ErrorM struct {
	Message string
}

type IssuedTicket struct {
	Plate string
	Days  []uint32
}

type Plate struct {
	Plate     string
	Timestamp uint32
}

type Tickets struct {
	mu      sync.Mutex
	Tickets []*Ticket
}

type Ticket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}

type Heartbeat struct {
	interval uint32
	ticker   *time.Ticker
}

type Dispatcher struct {
	NumRoads uint8
	Roads    []uint16
}

type Camera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
}

type Observation struct {
	Camera *Camera
	Plate  Plate
}
