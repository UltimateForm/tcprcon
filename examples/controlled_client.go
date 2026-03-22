package examples

import (
	"errors"
	"sync"
	"time"

	"github.com/UltimateForm/tcprcon/pkg/common_rcon"
	"github.com/UltimateForm/tcprcon/pkg/packet"
	"github.com/UltimateForm/tcprcon/pkg/rcon"
)

// ControlledClient wraps rcon.Client with mutex protection and command execution.
// This demonstrates how to safely use RCON in a concurrent context.
type ControlledClient struct {
	*rcon.Client
	lastUsed int64
	mu       sync.Mutex
}

// NewControlledClient creates a new controlled client connected to the given address.
func NewControlledClient(address string) (*ControlledClient, error) {
	baseClient, err := rcon.New(address)
	if err != nil {
		return nil, err
	}
	return &ControlledClient{
		Client:   baseClient,
		lastUsed: time.Now().Unix(),
	}, nil
}

// LastUsed returns the unix timestamp of the last time this client was used.
func (cc *ControlledClient) LastUsed() int64 {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.lastUsed
}

// Authenticate authenticates the client with the given password.
// All operations are protected by a mutex.
func (cc *ControlledClient) Authenticate(password string) (bool, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return common_rcon.Authenticate(cc, password)
}

// Execute sends a command and waits for the response with matching packet ID.
// Returns the response body as a string.
// Handles ID mismatches by continuing to read until the correct ID is found.
func (cc *ControlledClient) Execute(cmd string) (string, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	defer func() {
		cc.lastUsed = time.Now().Unix()
	}()

	writeId := cc.Id()

	// Send command packet
	execPacket := packet.New(writeId, packet.SERVERDATA_EXECCOMMAND, []byte(cmd))
	_, err := cc.Write(execPacket.Serialize())
	if err != nil {
		return "", errors.Join(errors.New("failed to write command"), err)
	}

	// Read response with timeout
	deadline := time.Now().Add(30 * time.Second)
	for {
		if time.Now().After(deadline) {
			return "", errors.New("timeout waiting for response")
		}
		cc.SetReadDeadline(time.Now().Add(10 * time.Second))

		responsePkt, err := packet.ReadWithId(cc, writeId)
		if errors.Is(err, packet.ErrPacketIdMismatch) {
			// Server sent us a broadcast or out-of-order packet; skip it
			continue
		}
		if err != nil {
			return "", errors.Join(errors.New("failed to read response"), err)
		}
		return responsePkt.BodyStr(), nil
	}
}
