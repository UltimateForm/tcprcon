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

	"github.com/UltimateForm/tcprcon/internal/logger"
	"github.com/UltimateForm/tcprcon/pkg/client"
	"github.com/UltimateForm/tcprcon/pkg/common"
	"golang.org/x/term"
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
			stdinbytes, _isPrefix, err := stdinread.ReadLine()
			if err != nil {
				return "", err
			}
			if _isPrefix {
				logger.Err.Println("prefix not supported")
				continue
			}
			r = string(stdinbytes)
		}
		if strings.ToLower(r) == "y" {
			password = envPassword
		}
	}
	if len(password) == 0 {
		fmt.Print("RCON PASSWORD: ")
		stdinbytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		password = string(stdinbytes)
	}
	return password, nil
}

func Execute() {
	flag.Parse()
	logLevel := uint8(logLevelParam)
	logger.Setup(logLevel)
	fullAddress := addressParam + ":" + strconv.Itoa(int(portParam))
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
	defer cancel()
	runRconTerminal(rcon, ctx, logLevel)
}
