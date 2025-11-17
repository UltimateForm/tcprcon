package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/UltimateForm/tcprcon/internal/client"
	"github.com/UltimateForm/tcprcon/internal/packet"
)

func main() {
	address := os.Getenv("rcon_address")
	port := os.Getenv("rcon_port")
	password := os.Getenv("rcon_password")
	log.Printf("Dialing %v at port %v\n", address, port)
	rcon, err := client.New(address + ":" + port)
	if err != nil {
		log.Fatal(err)
	}
	defer rcon.Close()

	log.Println("Building auth packet")
	authId := rcon.Id()
	authPacket := packet.NewAuthPacket(authId, password)
	log.Println("Sending auth packet")
	written, err := rcon.Write(authPacket.Serialize())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Written %v bytes of auth packet to connection", written)
	responsePkt, err := packet.Read(rcon, authId)
	if err != nil {
		log.Fatal(errors.Join(errors.New("authentication failed"), err))
	}
	if responsePkt.Type != packet.SERVERDATA_AUTH_RESPONSE {
		log.Fatal(
			fmt.Errorf(
				"unexpected packet type %v, expected %v",
				responsePkt.Type,
				packet.SERVERDATA_AUTH_RESPONSE,
			),
		)
	}
	for {
		log.Println("-----STARTING CMD EXCHANGE-----")
		stdinread := bufio.NewReader(os.Stdin)
		fmt.Print(">")
		cmd, _, err := stdinread.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		currId := rcon.Id()
		execPacket := packet.New(rcon.Id(), packet.SERVERDATA_EXECCOMMAND, cmd)
		rcon.Write(execPacket.Serialize())
		log.Println("Flushing writer...")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Reading from server...")
		responsePkt, err := packet.Read(rcon, currId)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("<%v\n", responsePkt.BodyStr())
	}

}
