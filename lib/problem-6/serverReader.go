package problem6

import (
	"errors"
	"protohackers/util"
)

const (
	ErrorType         byte = 0x10
	PlateType         byte = 0x20
	TicketType        byte = 0x21
	WantHeartbeatType byte = 0x40
	HearbeatType      byte = 0x41
	CameraType        byte = 0x80
	DispatchType      byte = 0x81
)

var (
	ErrUnknown = errors.New("unknown message")
)

type ServerWithReader struct {
	util.ServerTCP
}

// Reading an incoming message
func (s *ServerWithReader) ReadMessage(client *Client) (msg interface{}, err error) {
	typ, err := client.reader.r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch typ {
	case PlateType:
		plate := client.reader.readString()
		if plate == "" {
			return nil, nil
		}

		if client.camera == nil {
			// throw error
			er := ErrorM{
				Message: "Camera cannot be nil to observe plate",
			}
			er.writeError(client.writer.w)
			return
		}
		timestamp := client.reader.readUint32()

		msg = Plate{
			Plate:     plate,
			Timestamp: timestamp,
		}
	case DispatchType:
		if client.camera != nil || client.dispatcher != nil {
			// throw error
			er := ErrorM{
				Message: "Cannot have multiple personalities",
			}
			er.writeError(client.writer.w)
			return

		}

		numRoads, err := client.reader.r.ReadByte()
		if err != nil {
			return nil, err
		}
		roads := make([]uint16, numRoads)
		for i := 0; i < int(numRoads); i++ {
			roads = append(roads, client.reader.readUint16())
		}
		msg = Dispatcher{
			NumRoads: numRoads,
			Roads:    roads,
		}
	case WantHeartbeatType:
		if client.heartbeat != nil {
			// throw error
			er := ErrorM{
				Message: "Cannot have multiple heartbeats per connection",
			}
			er.writeError(client.writer.w)
			return

		}
		msg = Heartbeat{
			interval: client.reader.readUint32(),
		}
	case CameraType:
		if client.camera != nil || client.dispatcher != nil {
			// throw error
			er := ErrorM{
				Message: "Cannot have multiple personalities",
			}
			er.writeError(client.writer.w)
			return

		}
		msg = Camera{
			Road:  client.reader.readUint16(),
			Mile:  client.reader.readUint16(),
			Limit: client.reader.readUint16(),
		}
	default:
		return nil, ErrUnknown
	}
	return msg, err
}
