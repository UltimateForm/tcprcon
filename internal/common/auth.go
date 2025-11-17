package common

import (
	"fmt"
	"log"

	"github.com/UltimateForm/tcprcon/internal/client"
	"github.com/UltimateForm/tcprcon/internal/packet"
)

func Authenticate(rconClient *client.RCONClient, password string) (bool, error) {
	authId := rconClient.Id()
	authPacket := packet.NewAuthPacket(authId, password)
	written, err := rconClient.Write(authPacket.Serialize())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Written %v bytes of auth packet to connection", written)
	responsePkt, err := packet.Read(rconClient, authId)
	if err != nil {
		return false, err
	}
	if responsePkt.Type == packet.SERVERDATA_RESPONSE_VALUE {
		log.Printf("We got that flaky mythical empty server response, let's read again")
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
