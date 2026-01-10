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
	DisplayChannel chan string
	stdinChannel   chan byte
	fd             int
	prevState      *term.State
	cmdLine        []byte
	content        []string
}

func (src *app) Write(bytes []byte) (int, error) {
	src.DisplayChannel <- string(bytes)
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

func (src *app) DrawContent() error {
	_, height, err := term.GetSize(src.fd)
	if err != nil {
		return err
	}
	fmt.Print(ansi.ClearScreen + ansi.CursorHome)
	for i := range src.content {
		fmt.Print(src.content[i])
	}
	ansi.MoveCursorTo(height-1, 0)
	fmt.Print(">")
	fmt.Print(string(src.cmdLine))
	return nil
}
func (src *app) Run(context context.Context) error {

	// this could be an argument but i aint feeling yet
	src.fd = int(os.Stdin.Fd())
	if !term.IsTerminal(src.fd) {
		return errors.New("expected to run in terminal")
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

	go src.ListenStdin(context)
	for {
		select {
		case <-sigch:
			src.Close()
			return nil
		case <-context.Done():
			return nil
		case newStdinInput := <-src.stdinChannel:
			newCmd, isSubmission := constructCmdLine(newStdinInput, src.cmdLine)
			if isSubmission {
				src.DisplayChannel <- string(newCmd) + "\n\r"
				src.cmdLine = []byte{}
			} else {
				src.cmdLine = newCmd
			}
			if err := src.DrawContent(); err != nil {
				return err
			}
		case newDisplayInput := <-src.DisplayChannel:
			src.content = append(src.content, newDisplayInput)
			if err := src.DrawContent(); err != nil {
				return err
			}
		}
	}
}

func (src *app) Close() {
	term.Restore(src.fd, src.prevState)
}

func CreateApp() *app {
	// buffered, so we don't block on input
	displayChannel := make(chan string, 10)
	stdinChannel := make(chan byte)
	return &app{
		DisplayChannel: displayChannel,
		stdinChannel:   stdinChannel,
		content:        make([]string, 0),
	}
}
