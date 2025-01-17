package tcp

import "fmt"

type DataType byte

const (
	Message DataType = iota
	RGB8x8
	RGB16x16
	RGB8x32
	Mono8x8
)

type Packet struct {
	Version byte
	Type    DataType
	Data    []byte
}

func NewMessagePacket(data string) Packet {
	return Packet{
		Version: 1,
		Type:    Message,
		Data:    []byte(data),
	}
}

func NewLedPacket(dtype DataType, data []byte) Packet {
	return Packet{
		Version: 1,
		Type:    dtype,
		Data:    data,
	}
}

func (p Packet) Marshall() []byte {
	bytes := make([]byte, 0, 1+1+len(p.Data))
	bytes = append(bytes, p.Version, byte(p.Type))
	bytes = append(bytes, p.Data...)
	return bytes
}

func UnmarshallPacket(data []byte) (*Packet, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("data too small")
	}

	version := data[0]
	if version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	// TODO: trim data if not mesage
	return &Packet{
		Version: version,
		Type:    DataType(data[1]),
		Data:    data[2:],
	}, nil
}
