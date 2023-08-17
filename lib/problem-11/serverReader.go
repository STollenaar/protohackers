package problem11

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"protohackers/util"
)

const (
	HeaderLength = 5

	HelloMessageType     = 0x50
	ErrorMessageType     = 0x51
	OkMessageType        = 0x52
	DialAuthorityType    = 0x53
	TargetPopulationType = 0x54
	CreatePolicyType     = 0x55
	DeletePolicyType     = 0x56
	PolicyResultType     = 0x57
	SiteVisitType        = 0x58
)

type ServerWithReader struct {
	util.ServerTCP
}

func (ServerWithReader) ReadMessage(client *Client) (m Message, err error) {
	header := make([]byte, HeaderLength)
	_, err = io.ReadFull(client.reader.r, header)

	if err != nil {
		fmt.Println("Error reading header: ", err)
		return ErrorMessage{Message: "Error reading header"}, err
	}
	headerCopy := make([]byte, len(header))
	copy(headerCopy, header)

	headBuffer := bytes.NewBuffer(headerCopy)

	var typ uint8
	err = binary.Read(headBuffer, binary.BigEndian, &typ)
	if err != nil {
		fmt.Println("Error reading message type: ", err)
		return ErrorMessage{Message: "Error reading message type"}, err
	}

	var length uint32
	err = binary.Read(headBuffer, binary.BigEndian, &length)
	if err != nil {
		fmt.Println("Error reading message length: ", err)
		return ErrorMessage{Message: "Error reading message length"}, err
	}
	if length > 100000 {
		return ErrorMessage{Message: "Message too long"}, errors.New("message too long")
	}
	length -= HeaderLength // The message length includes the header. We already have the header so we can remove that length of it

	message := make([]byte, length)
	_, err = io.ReadFull(client.reader.r, message)
	if err != nil {
		fmt.Println("Error reading message: ", err)
		return ErrorMessage{Message: "Error reading message"}, err
	}

	chkBuf := bytes.NewBuffer(header)
	_, err = chkBuf.Write(message)
	if err != nil {
		fmt.Println("Error writing message to checksum buffer: ", err)
		return ErrorMessage{Message: "Error writing message to checksum buffer"}, err
	}

	if !ValidCheckSum(headBuffer.Bytes()) {
		fmt.Println("Error checksum not valid")
		return ErrorMessage{Message: "Error checksum not valid"}, err
	}
	message = message[:len(message)-1]
	m, err = Unmarshal(typ, message)
	return m, err
}

func sumData(data []byte) byte {
	sum := byte(0)
	for _, d := range data {
		sum += d
	}
	return sum
}

func ValidCheckSum(data []byte) bool {
	return sumData(data) == 0
}

func GenCheckSum(data []byte) {
	last := len(data) - 1
	data[last] = 0
	data[last] -= sumData(data)
}
