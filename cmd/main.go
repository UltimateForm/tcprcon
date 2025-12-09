package cmd

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/UltimateForm/tcprcon/internal/logger"
	"github.com/UltimateForm/tcprcon/pkg/client"
	"github.com/UltimateForm/tcprcon/pkg/common"
	"github.com/UltimateForm/tcprcon/pkg/packet"
)

var addressParam string
var portParam uint
var passwordParam string
var logLevelParam uint

func init() {
	flag.StringVar(&addressParam, "address", "localhost", "RCON address, excluding port")
	flag.UintVar(&portParam, "port", 7778, "RCON port")
	flag.StringVar(&passwordParam, "pw", "", "RCON password, if not provided will attempt to load from env variables, if unavailable will prompt")
	flag.UintVar(&logLevelParam, "log", logger.LevelWarning, "sets log level (syslog serverity tiers) for execution")
}

func determinePassword() (string, error) {
	if len(passwordParam) > 0 {
		logger.Debug.Println("using password from parameter")
		return passwordParam, nil
	}
	envPassword := os.Getenv("rcon_password")
	var password string
	if len(envPassword) > 0 {
		logger.Debug.Println("using password from os env")
		r := ""
		for r == "" {
			fmt.Print("RCON password found in environment variables, use for authentication? (y/n) ")
			stdinread := bufio.NewReader(os.Stdin)
			stdinbytes, _, err := stdinread.ReadLine()
			if err != nil {
				return "", err
			}
			r = string(stdinbytes)
		}
		if strings.ToLower(r) == "y" {
			password = envPassword
		}
	}
	if len(password) == 0 {
		return "", errors.New("unimplemented password retrieval path")
	}
	return password, nil
}

func Execute() {
	flag.Parse()
	logger.Setup(uint8(logLevelParam))
	fullAddress := addressParam + ":" + strconv.Itoa(int(portParam))
	shell := fmt.Sprintf("[rcon@%v]", fullAddress)
	password, err := determinePassword()
	if err != nil {
		logger.Critical.Fatal(err)
	}
	logger.Debug.Printf("Dialing %v at port %v\n", addressParam, portParam)
	rcon, err := client.New(fullAddress)
	if err != nil {
		logger.Critical.Fatal(err)
	}
	defer rcon.Close()

	logger.Debug.Println("Building auth packet")
	auhSuccess, authErr := common.Authenticate(rcon, password)
	if authErr != nil {
		logger.Err.Fatal(errors.Join(errors.New("auth failure"), authErr))
	}
	if !auhSuccess {
		logger.Err.Fatal(errors.New("unknown auth error"))
	}
	ctx, cancel := context.WithCancel(context.Background())
	packetChannel := packet.CreateResponseChannel(rcon, ctx)
	packetReader := func() {
		for {
			select {
			case <-ctx.Done():
				return
			case streamedPacket := <-packetChannel:
				if streamedPacket.Error != nil {
					logger.Err.Println(errors.Join(errors.New("error while reading from RCON client"), streamedPacket.Error))
				}
				fmt.Printf("OUT(%v): %v\n", streamedPacket.Id, streamedPacket.BodyStr())
			default:
				logger.Debug.Println("nothing to read, sleeping")
				time.Sleep(time.Duration(3 * time.Second))
			}
		}
	}
	go packetReader()
	for {
		logger.Info.Println("-----STARTING CMD EXCHANGE-----")
		stdinread := bufio.NewReader(os.Stdin)
		currId := rcon.Id()
		fmt.Printf("%v#(%v)", shell, currId)
		cmd, _, err := stdinread.ReadLine()
		if err != nil {
			logger.Critical.Fatal(err)
		}
		execPacket := packet.New(currId, packet.SERVERDATA_EXECCOMMAND, cmd)
		rcon.Write(execPacket.Serialize())
		if err != nil {
			cancel()
			logger.Critical.Fatal(err)
		}
		// logger.Debug.Println("Reading from server...")
		// responsePkt, err := packet.ReadWithId(rcon, currId)
		// if err != nil {
		// 	logger.Critical.Fatal(errors.Join(errors.New("error while reading from RCON client"), err))
		// }
		// streamedPacket := <-packetChannel
		// if streamedPacket.Error != nil {
		// 	cancel()
		// 	logger.Critical.Fatal(errors.Join(errors.New("error while reading from RCON client"), streamedPacket.Error))
		// }
		// fmt.Printf("OUT: %v\n", streamedPacket.BodyStr())
	}

}
