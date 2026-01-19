package common_rcon

import (
	"fmt"

	"github.com/UltimateForm/tcprcon/internal/logger"
	"github.com/UltimateForm/tcprcon/pkg/packet"
	"github.com/UltimateForm/tcprcon/pkg/rcon"
)

func Authenticate(rconClient *rcon.Client, password string) (bool, error) {
	authId := rconClient.Id()
	authPacket := packet.NewAuthPacket(authId, password)
	written, err := rconClient.Write(authPacket.Serialize())
	if err != nil {
		return false, err
	}
	logger.Debug.Printf("Written %v bytes of auth packet to connection", written)
	responsePkt, err := packet.ReadWithId(rconClient, authId)
	if err != nil {
		return false, err
	}

	if responsePkt.Type == packet.SERVERDATA_RESPONSE_VALUE {
		logger.Debug.Printf("We got that flaky mythical empty server response, let's read again")
		responsePkt, err = packet.ReadWithId(rconClient, authId)
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
