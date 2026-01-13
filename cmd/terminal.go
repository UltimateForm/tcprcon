package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/UltimateForm/tcprcon/internal/ansi"
	"github.com/UltimateForm/tcprcon/internal/fullterm"
	"github.com/UltimateForm/tcprcon/internal/logger"
	"github.com/UltimateForm/tcprcon/pkg/client"
	"github.com/UltimateForm/tcprcon/pkg/packet"
)

func runRconTerminal(client *client.RCONClient, ctx context.Context, logLevel uint8) {
	app := fullterm.CreateApp(fmt.Sprintf("rcon@%v", client.Address))
	logger.SetupCustomDestination(logLevel, app)
	errChan := make(chan error, 1)
	appRun := func() {
		errChan <- app.Run(ctx)
	}
	packetChannel := packet.CreateResponseChannel(client, ctx)
	packetReader := func() {
		for {
			select {
			case <-ctx.Done():
				return
			case streamedPacket := <-packetChannel:
				if streamedPacket.Error != nil {
					if errors.Is(streamedPacket.Error, os.ErrDeadlineExceeded) {
						logger.Debug.Println("read deadline reached; connection is idle or server is silent.")
						continue
					}
					logger.Err.Println(errors.Join(errors.New("error while reading from RCON client"), streamedPacket.Error))
				}
				fmt.Fprintf(
					app,
					"(%v): RESPONSE TYPE %v\n%v\n",
					ansi.Format(strconv.Itoa(int(streamedPacket.Id)), ansi.Green, ansi.Bold),
					ansi.Format(strconv.Itoa(int(streamedPacket.Type)), ansi.Green, ansi.Bold),
					ansi.Format(strings.TrimRight(streamedPacket.BodyStr(), "\n\r")+"\n", ansi.Green),
				)
			}
		}
	}
	submissionChan := app.Submissions()
	submissionReader := func() {
		for {
			select {
			case <-ctx.Done():
				return
			case cmd := <-submissionChan:
				execPacket := packet.New(client.Id(), packet.SERVERDATA_EXECCOMMAND, []byte(cmd))
				client.Write(execPacket.Serialize())
			}
		}
	}
	go submissionReader()
	go packetReader()
	go appRun()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			logger.Setup(logLevel)
			logger.Debug.Println("exiting app")
			if err != nil {
				logger.Critical.Println(err)
			}
			return
		}
	}
}
