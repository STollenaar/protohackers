package problem6

import (
	"math"
)

type PlateObserver struct {
	observer chan Observation
}

type plateObservation map[string][]Observation
type issueDates map[uint32]bool

// Making an observation
func (po *PlateObserver) makeObservation(camera *Camera, plate Plate) {
	po.observer <- Observation{
		Camera: camera,
		Plate:  plate,
	}
}

// Channel listener for when an observation happens
func (po *PlateObserver) listenForPlates() {
	roads := make(map[uint16]plateObservation)
	issuedTickets := make(map[string]issueDates)

	for {
		obs := <-po.observer

		road, ok := roads[obs.Camera.Road]
		if !ok {
			road = make(plateObservation)
			roads[obs.Camera.Road] = road
		}

		issuedTicket, ok := issuedTickets[obs.Plate.Plate]
		if !ok {
			issuedTicket = make(issueDates)
			issuedTickets[obs.Plate.Plate] = issuedTicket
		}

		tickets := checkObservations(road[obs.Plate.Plate], obs)

		for _, ticket := range tickets {
			day1 := ticket.Timestamp1 / 86400
			day2 := ticket.Timestamp2 / 86400

			if issuedForDays(issuedTicket, day1, day2) {
				continue
			}

			if client := getDispatcherClient(obs.Camera.Road); client != nil {
				// write to dispatcher
				ticket.writeTicket(client.writer.w)
			} else {
				pendingTickets.mu.Lock()
				pendingTickets.Tickets = append(pendingTickets.Tickets, &ticket)
				pendingTickets.mu.Unlock()
			}

			for i := day1; i <= day2; i++ {
				issuedTicket[i] = true
			}
		}
		road[obs.Plate.Plate] = append(road[obs.Plate.Plate], obs)
	}
}

// Get a dispatcher to send the ticket to
func getDispatcherClient(roadNmbr uint16) *Client {
	for _, client := range clients {
		if client.dispatcher != nil && client.dispatcher.hasRoad(roadNmbr) {
			return client
		}
	}
	return nil
}

// Checking if the current observation would make a valid ticket
func checkObservations(past []Observation, current Observation) (tickets []Ticket) {

	for _, p := range past {
		obv1, obv2 := p, current
		if obv1.Plate.Timestamp > obv2.Plate.Timestamp {
			obv1, obv2 = obv2, obv1
		}

		speed := calcAvgSpeed(p, current)

		if speed >= (float64(current.Camera.Limit) + 0.5) {
			// issue ticket
			tickets = append(tickets, Ticket{
				Plate:      current.Plate.Plate,
				Road:       current.Camera.Road,
				Mile1:      obv1.Camera.Mile,
				Timestamp1: obv1.Plate.Timestamp,
				Mile2:      obv2.Camera.Mile,
				Timestamp2: obv2.Plate.Timestamp,
				Speed:      uint16(speed * 100),
			})
		}
	}
	return tickets
}

// Calculating the average speed between 2 observations
func calcAvgSpeed(obv1, obv2 Observation) float64 {
	distance := math.Abs(float64(obv1.Camera.Mile) - float64(obv2.Camera.Mile))
	time := math.Abs(float64(obv1.Plate.Timestamp) - float64(obv2.Plate.Timestamp))

	speed := (distance / time) * 3600
	return speed
}

// Checking if the ticket was already issued for the spanned days
func issuedForDays(issuedTicket issueDates, day1, day2 uint32) bool {
	for i := day1; i <= day2; i++ {
		if issuedTicket[i] {
			return true
		}
	}
	return false
}
