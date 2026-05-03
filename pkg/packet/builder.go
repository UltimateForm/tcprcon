package packet

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/UltimateForm/tcprcon/pkg/logger"
)

type RCONPacket struct {
	Id   int32
	Type int32
	Body []byte
}

func (src RCONPacket) BodyStr() string {
	return string(src.Body)
}

func New(id int32, pktType int32, body []byte) RCONPacket {
	return RCONPacket{
		Id:   id,
		Type: pktType,
		Body: body,
	}
}

func NewAuthPacket(id int32, password string) RCONPacket {
	return New(
		id, SERVERDATA_AUTH, []byte(password),
	)
}

func (src RCONPacket) Serialize() []byte {
	size := 8 + len(src.Body) + 2
	logger.Debug.Printf("Building packet of size %v\n", size)
	var bytesSlice []byte = make([]byte, size+4)
	binary.LittleEndian.PutUint32(bytesSlice[0:4], uint32(size))
	binary.LittleEndian.PutUint32(bytesSlice[4:8], uint32(src.Id))
	binary.LittleEndian.PutUint32(bytesSlice[8:12], uint32(src.Type))
	copy(bytesSlice[12:], src.Body)
	bytesSlice[12+len(src.Body)] = 0
	bytesSlice[12+len(src.Body)+1] = 0
	logger.Debug.Printf("Final packet length %v", len(bytesSlice))
	return bytesSlice
}

func readPacket(reader io.Reader) (RCONPacket, error) {
	dword := make([]byte, 4)
	_, err := io.ReadFull(reader, dword)
	if err != nil {
		return RCONPacket{}, err
	}
	packetSize := binary.LittleEndian.Uint32(dword)
	packetBytes := make([]byte, packetSize)
	_, err = io.ReadFull(reader, packetBytes)
	if err != nil {
		return RCONPacket{}, err
	}
	id := int32(binary.LittleEndian.Uint32(packetBytes[0:4]))
	packetType := int32(binary.LittleEndian.Uint32(packetBytes[4:8]))
	body := bytes.TrimRight(packetBytes[8:], "\x00")
	return New(id, packetType, body), nil
}

func ReadWithId(reader io.Reader, expectedId int32) (RCONPacket, error) {
	pkt, err := readPacket(reader)
	if err != nil {
		return pkt, err
	}
	if pkt.Id != expectedId {
		return pkt, ErrPacketIdMismatch
	}
	return pkt, nil
}

func Read(reader io.Reader) (RCONPacket, error) {
	return readPacket(reader)
}
