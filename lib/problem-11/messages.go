package problem11

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Message interface {
	Type() uint8
	Unmarshal(d []byte) (Message, error)
	Marshal(client *bufio.Writer)
}

func Unmarshal(kind uint8, data []byte) (message Message, e error) {
	switch kind {
	case HelloMessageType:
		h := HelloMessage{}
		return h.Unmarshal(data)
	case OkMessageType:
		o := OkMessage{}
		return o.Unmarshal(data)
	case DialAuthorityType:
		d := DialAuthorityMessage{}
		return d.Unmarshal(data)
	case TargetPopulationType:
		t := TargetPopulationMessage{}
		return t.Unmarshal(data)
	case CreatePolicyType:
		c := CreatePolicyMessage{}
		return c.Unmarshal(data)
	case DeletePolicyType:
		d := DeletePolicyMessage{}
		return d.Unmarshal(data)
	case PolicyResultType:
		p := PolicyResultMessage{}
		return p.Unmarshal(data)
	case SiteVisitType:
		s := SiteVisitMessage{}
		return s.Unmarshal(data)
	default:
		e := ErrorMessage{}
		return e.Unmarshal(data)
	}
}

type HelloMessage struct {
	Protocol string
	Version  uint32
}

func (m HelloMessage) Type() uint8 {
	return HelloMessageType
}

func (m HelloMessage) Unmarshal(data []byte) (Message, error) {
	reader := bytes.NewReader(data)

	var protLength, version uint32
	err := binary.Read(reader, binary.BigEndian, &protLength)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}
	protocol := make([]byte, protLength)
	err = binary.Read(reader, binary.BigEndian, protocol)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	err = binary.Read(reader, binary.BigEndian, &version)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}
	if string(protocol) != "pestcontrol" || version != 1 {
		message := fmt.Errorf("unknown protocol %s, %d", string(protocol), version)
		return ErrorMessage{Message: message.Error()}, err
	}

	return HelloMessage{Protocol: string(protocol), Version: version}, nil
}

func (m HelloMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	writeUint32(buffer, uint32(1+4+4+len(m.Protocol)+4+1)) // Type, total Length (4bit), length of string (4bit), string, version (4bit), chksm
	writeString(buffer, m.Protocol)
	writeUint32(buffer, m.Version)
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type ErrorMessage struct {
	Message string
}

func (m ErrorMessage) Type() uint8 {
	return ErrorMessageType
}

func (m ErrorMessage) Unmarshal(data []byte) (Message, error) {
	return ErrorMessage{Message: string(data)}, nil
}

func (m ErrorMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)
	writeByte(buffer, m.Type())
	writeUint32(buffer, uint32(1+4+4+len(m.Message)+1)) // Type, total Length (4bit), length of string (4bit), string, chksm
	writeString(buffer, m.Message)
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	
	client.Write(buffer.Bytes())
	client.Flush()
}

type OkMessage struct{}

func (m OkMessage) Type() uint8 {
	return OkMessageType
}

func (m OkMessage) Unmarshal(data []byte) (Message, error) {
	return OkMessage{}, nil
}
func (m OkMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	writeUint32(buffer, 6)
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type DialAuthorityMessage struct {
	site uint32
}

func (m DialAuthorityMessage) Type() uint8 {
	return DialAuthorityType
}

func (m DialAuthorityMessage) Unmarshal(data []byte) (Message, error) {
	reader := bytes.NewReader(data)

	var site uint32
	err := binary.Read(reader, binary.BigEndian, &site)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}
	return DialAuthorityMessage{site: site}, nil
}

func (m DialAuthorityMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	writeUint32(buffer, 1+4+4+1)
	writeUint32(buffer, m.site)
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type MinMax struct {
	Min uint32 `json:"min"`
	Max uint32 `json:"max"`
}

type PopulationTarget map[string]MinMax

type TargetPopulationMessage struct {
	Site        uint32           `json:"site"`
	Populations PopulationTarget `json:"populations"`
}

func (m TargetPopulationMessage) Type() uint8 {
	return TargetPopulationType
}

func (m TargetPopulationMessage) Unmarshal(data []byte) (Message, error) {
	var site, length uint32
	populations := make(PopulationTarget)

	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.BigEndian, &site)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	err = binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	for len(populations) < int(length) {
		var speciesLength, min, max uint32
		err = binary.Read(reader, binary.BigEndian, &speciesLength)
		if err != nil {
			fmt.Printf("Error reading length: %s\n", err)
			return ErrorMessage{Message: err.Error()}, err
		}

		species := make([]byte, speciesLength)
		err = binary.Read(reader, binary.BigEndian, species)
		if err != nil {
			fmt.Printf("Error reading species: %s\n", err)

			return ErrorMessage{Message: err.Error()}, err
		}
		err = binary.Read(reader, binary.BigEndian, &min)
		if err != nil {
			fmt.Printf("Error reading minimum: %s\n", err)
			return ErrorMessage{Message: err.Error()}, err
		}

		err = binary.Read(reader, binary.BigEndian, &max)
		if err != nil {
			fmt.Printf("Error reading maximum: %s\n", err)
			return ErrorMessage{Message: err.Error()}, err
		}

		if min > max {
			t := min
			min = max
			max = t
		}
		populations[string(species)] = MinMax{Min: min, Max: max}
	}
	return TargetPopulationMessage{Site: site, Populations: populations}, nil
}

func (m TargetPopulationMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	var specLength int
	for k := range m.Populations {
		specLength += len(k)
	}
	writeUint32(buffer, uint32(1+4+1+4+(4*len(m.Populations)*3)+specLength))
	writeUint32(buffer, m.Site)
	writeUint32(buffer, uint32(len(m.Populations)))
	for k, p := range m.Populations {
		writeString(buffer, k)
		writeUint32(buffer, p.Min)
		writeUint32(buffer, p.Max)
	}
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type ActionType byte

const (
	CullAction     ActionType = 0x90
	ConserveAction ActionType = 0xa0
)

type CreatePolicyMessage struct {
	species string
	action  ActionType
}

func (m CreatePolicyMessage) Type() uint8 {
	return CreatePolicyType
}
func (m CreatePolicyMessage) Unmarshal(data []byte) (Message, error) {
	reader := bytes.NewReader(data)

	var speciesLength uint32
	err := binary.Read(reader, binary.BigEndian, &speciesLength)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	species := make([]byte, speciesLength)
	err = binary.Read(reader, binary.BigEndian, species)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	var action ActionType
	err = binary.Read(reader, binary.BigEndian, &action)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}
	if action != CullAction && action != ConserveAction {
		message := errors.New("unknown action")
		return ErrorMessage{Message: message.Error()}, message
	}

	return CreatePolicyMessage{species: string(species), action: action}, nil
}

func (m CreatePolicyMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	writeUint32(buffer, uint32(1+4+4+len(m.species)+1+1))
	writeString(buffer, m.species)
	writeByte(buffer, byte(m.action))
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type DeletePolicyMessage struct {
	policy uint32
}

func (m DeletePolicyMessage) Type() uint8 {
	return DeletePolicyType
}

func (m DeletePolicyMessage) Unmarshal(data []byte) (Message, error) {
	reader := bytes.NewReader(data)

	var policy uint32
	err := binary.Read(reader, binary.BigEndian, &policy)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	return DeletePolicyMessage{policy: policy}, nil
}

func (m DeletePolicyMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	writeUint32(buffer, 10)
	writeUint32(buffer, m.policy)
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type PolicyResultMessage struct {
	policy uint32
}

func (m PolicyResultMessage) Type() uint8 {
	return PolicyResultType
}

func (m PolicyResultMessage) Unmarshal(data []byte) (Message, error) {
	reader := bytes.NewReader(data)

	var policy uint32
	err := binary.Read(reader, binary.BigEndian, &policy)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	return PolicyResultMessage{policy: policy}, nil

}
func (m PolicyResultMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	writeUint32(buffer, 10)
	writeUint32(buffer, m.policy)
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}

type PopulationVisit map[string]uint32

type SiteVisitMessage struct {
	site        uint32
	populations PopulationVisit
}

func (m SiteVisitMessage) Type() uint8 {
	return SiteVisitType
}

func (m SiteVisitMessage) Unmarshal(data []byte) (Message, error) {
	var site, length uint32
	populations := make(PopulationVisit)

	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.BigEndian, &site)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	err = binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return ErrorMessage{Message: err.Error()}, err
	}

	for i := 0; i < int(length); i++ {

		var speciesLength, count uint32
		err = binary.Read(reader, binary.BigEndian, &speciesLength)
		if err != nil {
			return ErrorMessage{Message: err.Error()}, err
		}

		species := make([]byte, speciesLength)
		err = binary.Read(reader, binary.BigEndian, &species)
		if err != nil {
			return ErrorMessage{Message: err.Error()}, err
		}
		err = binary.Read(reader, binary.BigEndian, &count)
		if err != nil {
			return ErrorMessage{Message: err.Error()}, err
		}

		if c, ok := populations[string(species)]; ok && c != count {
			message := fmt.Sprintf("conflicting count for species %s, being initial count: %d, and now %d", string(species), c, count)
			return ErrorMessage{Message: message}, errors.New(message)
		}
		populations[string(species)] = count
	}

	return SiteVisitMessage{site: site, populations: populations}, nil
}

func (m SiteVisitMessage) Marshal(client *bufio.Writer) {
	buffer := new(bytes.Buffer)

	writeByte(buffer, m.Type())
	specLength := len(m.populations)

	writeUint32(buffer, uint32(1+4+1+4+(4*len(m.populations)*2)+specLength))
	writeUint32(buffer, m.site)
	writeUint32(buffer, uint32(len(m.populations)))
	for species, count := range m.populations {
		writeString(buffer, species)
		writeUint32(buffer, count)
	}
	writeByte(buffer, 0)
	GenCheckSum(buffer.Bytes())
	client.Write(buffer.Bytes())
	client.Flush()
}
