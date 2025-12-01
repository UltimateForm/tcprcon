package client

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// MockConn implements net.Conn for testing
type MockConn struct {
	readData  []byte
	writeData []byte
	readPos   int
	closed    bool
}

func (m *MockConn) Read(p []byte) (n int, err error) {
	if m.readPos >= len(m.readData) {
		return 0, nil
	}
	n = copy(p, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *MockConn) Write(p []byte) (n int, err error) {
	m.writeData = append(m.writeData, p...)
	return len(p), nil
}

func (m *MockConn) Close() error {
	m.closed = true
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestRCONClientId(t *testing.T) {
	mock := &MockConn{}
	client := &RCONClient{
		Address: "test:27015",
		con:     mock,
		count:   42,
	}

	id := client.Id()
	if id != 42 {
		t.Fatalf("Id mismatch: got %d want 42", id)
	}
}

func TestRCONClientWrite(t *testing.T) {
	mock := &MockConn{}
	client := &RCONClient{
		Address: "test:27015",
		con:     mock,
		count:   0,
	}

	data := []byte("test packet")
	n, err := client.Write(data)

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Fatalf("bytes written mismatch: got %d want %d", n, len(data))
	}

	if !bytes.Equal(mock.writeData, data) {
		t.Fatalf("written data mismatch: got %v want %v", mock.writeData, data)
	}

	// Check that count was incremented
	if client.count != 1 {
		t.Fatalf("count should increment after write: got %d want 1", client.count)
	}
}

func TestRCONClientWriteIncrementsCount(t *testing.T) {
	mock := &MockConn{}
	client := &RCONClient{
		Address: "test:27015",
		con:     mock,
		count:   0,
	}

	// Write multiple times and verify count increments
	for i := 0; i < 5; i++ {
		client.Write([]byte("data"))
		if client.count != int32(i+1) {
			t.Fatalf("count after write %d: got %d want %d", i+1, client.count, i+1)
		}
	}
}

func TestRCONClientRead(t *testing.T) {
	testData := []byte("response data")
	mock := &MockConn{readData: testData}

	client := &RCONClient{
		Address: "test:27015",
		con:     mock,
		count:   0,
	}

	p := make([]byte, len(testData))
	n, err := client.Read(p)

	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != len(testData) {
		t.Fatalf("bytes read mismatch: got %d want %d", n, len(testData))
	}

	if !bytes.Equal(p, testData) {
		t.Fatalf("read data mismatch: got %v want %v", p, testData)
	}

	// Verify count was NOT incremented (Read doesn't increment)
	if client.count != 0 {
		t.Fatalf("count should not increment on Read: got %d want 0", client.count)
	}
}
