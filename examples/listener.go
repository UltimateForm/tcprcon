package examples

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"github.com/UltimateForm/tcprcon/pkg/packet"
)

const (
	listenerChannelBuffer = 32
	keepaliveIntervalSecs = 100
	reconnectDelaySecs    = 5
)

// EventListener demonstrates how to use CreateResponseChannel to stream
// and handle asynchronous server events.
type EventListener struct {
	client      *ControlledClient
	uri         string
	password    string
	Events      <-chan string // Generic event channel
	eventsCh    chan string
	logger      *log.Logger
}

// NewEventListener creates a listener connected to the RCON server.
func NewEventListener(uri, password string) (*EventListener, error) {
	client, err := NewControlledClient(uri)
	if err != nil {
		return nil, err
	}
	ok, err := client.Authenticate(password)
	if err != nil {
		client.Close()
		return nil, err
	}
	if !ok {
		client.Close()
		return nil, errors.New("authentication failed")
	}

	l := &EventListener{
		client:   client,
		uri:      uri,
		password: password,
		eventsCh: make(chan string, listenerChannelBuffer),
		logger: log.New(
			log.Default().Writer(),
			"[EventListener] ",
			log.Default().Flags(),
		),
	}
	l.Events = l.eventsCh
	return l, nil
}

// reconnect closes the current connection and establishes a new one.
func (l *EventListener) reconnect() error {
	l.client.Close()
	client, err := NewControlledClient(l.uri)
	if err != nil {
		return err
	}
	ok, err := client.Authenticate(l.password)
	if err != nil {
		client.Close()
		return err
	}
	if !ok {
		client.Close()
		return errors.New("authentication failed")
	}
	l.client = client
	l.logger.Println("reconnected successfully")
	return nil
}

// stream continuously reads packets from the server and sends them to the event channel.
// On connection loss, it attempts to reconnect.
func (l *EventListener) stream(ctx context.Context) {
	for {
		connCtx, cancelConn := context.WithCancel(ctx)
		go l.keepalive(connCtx)

		packetChan := packet.CreateResponseChannel(l.client, connCtx)
		for pkt := range packetChan {
			if pkt.Error != nil {
				if netErr, ok := pkt.Error.(net.Error); ok && netErr.Timeout() {
					continue
				}
				l.logger.Printf("stream error: %v", pkt.Error)
				break
			}
			body := pkt.BodyStr()
			if body == "Keeping client alive" {
				continue
			}
			select {
			case l.eventsCh <- body:
			default:
				l.logger.Println("event channel full, dropping event")
			}
		}

		cancelConn()

		if ctx.Err() != nil {
			return
		}

		l.logger.Println("connection lost, reconnecting...")
		for {
			time.Sleep(reconnectDelaySecs * time.Second)
			if err := l.reconnect(); err != nil {
				l.logger.Printf("reconnect failed: %v, retrying...", err)
				continue
			}
			break
		}
	}
}

// keepalive periodically sends a heartbeat command to keep the connection alive.
func (l *EventListener) keepalive(ctx context.Context) {
	ticker := time.NewTicker(keepaliveIntervalSecs * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := l.client.Execute("alive")
			if err != nil {
				l.logger.Printf("keepalive error: %v", err)
			}
		}
	}
}

// Run starts the listener in a background goroutine.
// Events are sent to the Events channel.
func (l *EventListener) Run(ctx context.Context) {
	go l.stream(ctx)
}

// Close stops the listener and closes the connection.
func (l *EventListener) Close() error {
	return l.client.Close()
}
