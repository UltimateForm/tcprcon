package fullterm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type app struct {
	// this is not stdin, just stuff to draw, silly goose
	InputChannel chan string
	fd           int
	prevState    *term.State
}

func (src *app) Write(bytes []byte) (int, error) {
	src.InputChannel <- string(bytes)
	return len(bytes), nil
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
	currFlags.Iflag |= unix.ICRNL
	currFlags.Lflag |= (unix.ISIG | unix.ECHO | unix.ECHONL | unix.ICANON | unix.ECHONL | unix.ECHOE)

	// fyi there's a TCSETS as well that applies the setting differently
	if err := unix.IoctlSetTermios(src.fd, unix.TCSETSF, currFlags); err != nil {
		return err
	}

	content := make([]string, 0)
	for {
		select {
		case <-sigch:
			src.Close()
			return nil
		case <-context.Done():
			return nil
		case newInput := <-src.InputChannel:
			content := append(content, newInput)
			fmt.Print("\033[2J\033[H")
			for i := range content {
				fmt.Print(i)
			}
			fmt.Print(newInput)
			fmt.Printf("\033[%v;0H>", height-1)
		}
	}
}

func (src *app) Close() {
	term.Restore(src.fd, src.prevState)
}

func CreateApp() *app {
	inputChannel := make(chan string)
	return &app{
		InputChannel: inputChannel,
	}
}
