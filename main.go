package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/UltimateForm/tcprcon/internal/logger"
	"github.com/UltimateForm/tcprcon/pkg/client"
	"github.com/UltimateForm/tcprcon/pkg/common"
	"github.com/UltimateForm/tcprcon/pkg/packet"
)

func main() {
	address := os.Getenv("rcon_address")
	port := os.Getenv("rcon_port")
	password := os.Getenv("rcon_password")
	logger.Debug.Printf("Dialing %v at port %v\n", address, port)
	fullAddress := address + ":" + port
	shell := fmt.Sprintf("[rcon@%v]", fullAddress)
	rcon, err := client.New(fullAddress)
	if err != nil {
		logger.Critical.Fatal(err)
	}
	defer rcon.Close()

	logger.Debug.Println("Building auth packet")
	auhSuccess, authErr := common.Authenticate(rcon, password)
	if authErr != nil {
		logger.Err.Fatal(err)
	}
	if !auhSuccess {
		logger.Err.Fatal(errors.New("auth failure"))
	}
	for {
		logger.Info.Println("-----STARTING CMD EXCHANGE-----")
		stdinread := bufio.NewReader(os.Stdin)
		fmt.Printf("%v#", shell)
		cmd, _, err := stdinread.ReadLine()
		if err != nil {
			logger.Critical.Fatal(err)
		}
		currId := rcon.Id()
		execPacket := packet.New(currId, packet.SERVERDATA_EXECCOMMAND, cmd)
		rcon.Write(execPacket.Serialize())
		logger.Debug.Println("Flushing writer...")
		if err != nil {
			logger.Critical.Fatal(err)
		}
		logger.Debug.Println("Reading from server...")
		responsePkt, err := packet.Read(rcon, currId)
		if err != nil {
			logger.Critical.Fatal(errors.Join(errors.New("error while reading from RCON client"), err))
		}
		fmt.Printf("OUT: %v\n", responsePkt.BodyStr())
	}

}
