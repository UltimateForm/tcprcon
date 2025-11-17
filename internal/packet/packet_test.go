package packet

import (
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
