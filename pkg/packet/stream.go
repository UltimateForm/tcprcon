package packet

import (
	"context"
	"io"
	"time"
)

type StreamedPacket struct {
	Error error
	RCONPacket
}

type responseConn interface {
	io.Reader
	SetReadDeadline(t time.Time) error
}

func CreateResponseChannel(con responseConn, ctx context.Context) <-chan StreamedPacket {
	packetChan := make(chan StreamedPacket)
	stream := func() {
		defer close(packetChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				con.SetReadDeadline(time.Now().Add(60 * time.Second))
				packet, err := Read(con)
				packetChan <- StreamedPacket{
					Error:      err,
					RCONPacket: packet,
				}
			}

		}
	}
	go stream()
	return packetChan
}
