package common

import (
	"fmt"

	"github.com/UltimateForm/tcprcon/internal/logger"
	"github.com/UltimateForm/tcprcon/pkg/client"
	"github.com/UltimateForm/tcprcon/pkg/packet"
)

func Authenticate(rconClient *client.RCONClient, password string) (bool, error) {
	authId := rconClient.Id()
	authPacket := packet.NewAuthPacket(authId, password)
	written, err := rconClient.Write(authPacket.Serialize())
	if err != nil {
		logger.Critical.Fatal(err)
	}
	logger.Debug.Printf("Written %v bytes of auth packet to connection", written)
	responsePkt, err := packet.Read(rconClient, authId)
	if err != nil {
		return false, err
	}
	if responsePkt.Type == packet.SERVERDATA_RESPONSE_VALUE {
		logger.Debug.Println("We got that flaky mythical empty server response, let's read again")
		responsePkt, err = packet.Read(rconClient, authId)
		if err != nil {
			return false, err
		}
	}
	if responsePkt.Type != packet.SERVERDATA_AUTH_RESPONSE {
		return false, fmt.Errorf(
			"unexpected packet type %v, expected %v",
			responsePkt.Type,
			packet.SERVERDATA_AUTH_RESPONSE,
		)
	}
	return true, nil
}
