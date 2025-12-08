package packet

import (
	"context"
	"time"

	"github.com/UltimateForm/tcprcon/pkg/client"
)

type StreamedPacket struct {
	Error error
	RCONPacket
}

func CreateResponseChannel(con *client.RCONClient, ctx context.Context) <-chan StreamedPacket {
	packetChan := make(chan StreamedPacket)
	stream := func() {
		defer close(packetChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				con.SetReadDeadline(time.Now().Add(10 * time.Second))
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
