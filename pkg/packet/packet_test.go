package packet

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSerializePacket(t *testing.T) {
	// Basic values for the test
	id := int32(0x11223344)
	ptype := int32(0x55667788)
	body := []byte("hello")
	pktData := RCONPacket{
		Id:   id,
		Body: body,
		Type: ptype,
	}
	pkt := pktData.Serialize()

	// size is 8 (id+type) + len(body) + 2 (two null terminators)
	expectedSize := uint32(8 + len(body) + 2)

	if len(pkt) != int(expectedSize)+4 {
		t.Fatalf("packet length mismatch: got %d want %d", len(pkt), int(expectedSize)+4)
	}

	// Verify size field (first 4 bytes)
	gotSize := binary.LittleEndian.Uint32(pkt[0:4])
	if gotSize != expectedSize {
		t.Fatalf("size field mismatch: got %d want %d", gotSize, expectedSize)
	}

	// Verify id and packet type
	gotId := int32(binary.LittleEndian.Uint32(pkt[4:8]))
	if gotId != id {
		t.Fatalf("id mismatch: got 0x%x want 0x%x", uint32(gotId), uint32(id))
	}

	gotType := int32(binary.LittleEndian.Uint32(pkt[8:12]))
	if gotType != ptype {
		t.Fatalf("type mismatch: got 0x%x want 0x%x", uint32(gotType), uint32(ptype))
	}

	// Verify body bytes
	bodyStart := 12
	bodyEnd := bodyStart + len(body)
	if string(pkt[bodyStart:bodyEnd]) != string(body) {
		t.Fatalf("body mismatch: got %v want %v", pkt[bodyStart:bodyEnd], body)
	}

	// Verify two null terminators after body
	if pkt[bodyEnd] != 0 {
		t.Fatalf("first null terminator not zero: got 0x%x", pkt[bodyEnd])
	}
	if pkt[bodyEnd+1] != 0 {
		t.Fatalf("second null terminator not zero: got 0x%x", pkt[bodyEnd+1])
	}
}

func TestReadPacket(t *testing.T) {
	// i could use the existing packet.New() command but idk i wanna make sure my tests are not too sticky
	id := int32(42)
	ptype := int32(SERVERDATA_AUTH_RESPONSE)
	body := []byte("server response body")

	// keep sync'd with internal/packet/builder.go Serialize() func
	size := uint32(8 + len(body) + 2)
	packet := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(packet[0:4], size)
	binary.LittleEndian.PutUint32(packet[4:8], uint32(id))
	binary.LittleEndian.PutUint32(packet[8:12], uint32(ptype))
	copy(packet[12:], body)
	packet[12+len(body)] = 0
	packet[12+len(body)+1] = 0

	reader := bytes.NewReader(packet)
	pkt, err := ReadWithId(reader, id)

	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if pkt.Id != id {
		t.Fatalf("id mismatch: got %d want %d", pkt.Id, id)
	}

	if pkt.Type != ptype {
		t.Fatalf("type mismatch: got %d want %d", pkt.Type, ptype)
	}

	expectedBodyStr := string(body)
	receivedBodyStr := pkt.BodyStr()
	if expectedBodyStr != receivedBodyStr {
		t.Fatalf("body mismatch: got %q want %q", receivedBodyStr, expectedBodyStr)
	}
}

func TestReadPacketIdMismatch(t *testing.T) {
	id := int32(42)
	expectedId := int32(99)
	ptype := int32(SERVERDATA_AUTH_RESPONSE)
	body := []byte("test")

	size := uint32(8 + len(body) + 2)
	packet := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(packet[0:4], size)
	binary.LittleEndian.PutUint32(packet[4:8], uint32(id))
	binary.LittleEndian.PutUint32(packet[8:12], uint32(ptype))
	copy(packet[12:], body)

	reader := bytes.NewReader(packet)
	pkt, err := ReadWithId(reader, expectedId)

	if err != ErrPacketIdMismatch {
		t.Fatalf("expected ErrPacketIdMismatch, got %v", err)
	}

	if pkt.Id != id {
		t.Fatalf("packet id should still be set: got %d want %d", pkt.Id, id)
	}
}

func TestReadPacketEmptyBody(t *testing.T) {
	id := int32(1)
	ptype := int32(SERVERDATA_RESPONSE_VALUE)
	body := []byte{}

	size := uint32(8 + len(body) + 2)
	packet := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(packet[0:4], size)
	binary.LittleEndian.PutUint32(packet[4:8], uint32(id))
	binary.LittleEndian.PutUint32(packet[8:12], uint32(ptype))
	packet[12] = 0
	packet[13] = 0

	reader := bytes.NewReader(packet)
	pkt, err := ReadWithId(reader, id)

	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(pkt.Body) != 0 {
		t.Fatalf("body should be empty: got %v", pkt.Body)
	}
}
