package packet

import (
	"encoding/binary"
	"log"
)

func BuildPacket(id int32, packetType int32, bodyBytes []byte) []byte {
	size := 8 + len(bodyBytes) + 2
	log.Printf("Building packet of size %v\n", size)
	var bytesSlice []byte = make([]byte, size+4)
	binary.LittleEndian.PutUint32(bytesSlice[0:4], uint32(size))
	binary.LittleEndian.PutUint32(bytesSlice[4:8], uint32(id))
	binary.LittleEndian.PutUint32(bytesSlice[8:12], uint32(packetType))
	copy(bytesSlice[12:], bodyBytes)
	bytesSlice[12+len(bodyBytes)] = 0
	bytesSlice[12+len(bodyBytes)+1] = 0
	log.Printf("Final packet length %v", len(bytesSlice))
	return bytesSlice
}
