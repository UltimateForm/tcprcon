package fullterm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/UltimateForm/tcprcon/internal/ansi"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type app struct {
	// this is not stdin, just stuff to draw, silly goose
	InputChannel chan string
	stdinChannel chan byte
	fd           int
	prevState    *term.State
	cmdLine      []byte
}

func (src *app) Write(bytes []byte) (int, error) {
	src.InputChannel <- string(bytes)
	return len(bytes), nil
}

func (src *app) ListenStdin(context context.Context) {
	for {
		select {
		case <-context.Done():
			return
		default:
			b := make([]byte, 1)
			_, err := os.Stdin.Read(b)
			if err != nil {
				return
			}
			src.stdinChannel <- b[0]
		}
	}
}

func constructCmdLine(newByte byte, cmdLine []byte) ([]byte, bool) {
	isSubmission := false
	switch newByte {
	case 127, 8: // backspace, delete
		if len(cmdLine) > 0 {
			cmdLine = cmdLine[:len(cmdLine)-1]
		}
	case 13, 10: // enter
		isSubmission = true
	default:
		cmdLine = append(cmdLine, newByte)
	}
	return cmdLine, isSubmission
}

func (src *app) Run(context context.Context) error {

	// this could be an argument but i aint feeling yet
	src.fd = int(os.Stdin.Fd())
	if !term.IsTerminal(src.fd) {
		return errors.New("expected to run in terminal")
	}
	_, height, err := term.GetSize(src.fd)
	if err != nil {
		return err
	}

	prevState, err := term.MakeRaw(src.fd)
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGABRT)

	if err != nil {
		return err
	}
	src.prevState = prevState
	defer src.Close()

	currFlags, err := unix.IoctlGetTermios(src.fd, unix.TCGETS)
	if err != nil {
		return err
	}
	currFlags.Lflag |= unix.ISIG

	// fyi there's a TCSETS as well that applies the setting differently
	if err := unix.IoctlSetTermios(src.fd, unix.TCSETSF, currFlags); err != nil {
		return err
	}

	content := make([]string, 0)
	go src.ListenStdin(context)
	for {
		select {
		case <-sigch:
			src.Close()
			return nil
		case <-context.Done():
			return nil
		case newInput := <-src.stdinChannel:
			fmt.Print(ansi.ClearScreen + ansi.CursorHome)
			for i := range content {
				fmt.Print(content[i])
			}
			ansi.MoveCursorTo(height-1, 0)
			fmt.Print(">")
			newCmd, isSubmission := constructCmdLine(newInput, src.cmdLine)
			if isSubmission {
				src.InputChannel <- string(newCmd) + "\n\r"
				src.cmdLine = []byte{}
			} else {
				src.cmdLine = newCmd
				fmt.Print(string(newCmd))
			}
		case newInput := <-src.InputChannel:
			content = append(content, newInput)
			fmt.Print(ansi.ClearScreen + ansi.CursorHome)
			for i := range content {
				fmt.Print(content[i])
			}
			ansi.MoveCursorTo(height-1, 0)
			fmt.Print(">")
			fmt.Print(string(src.cmdLine))
		}
	}
}

func (src *app) Close() {
	term.Restore(src.fd, src.prevState)
}

func CreateApp() *app {
	// buffered, so we don't block on input
	inputChannel := make(chan string, 10)
	stdinChannel := make(chan byte)
	return &app{
		InputChannel: inputChannel,
		stdinChannel: stdinChannel,
	}
}
