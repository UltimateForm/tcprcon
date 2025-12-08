package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func draw(counter int, height int, lines [][]byte) {

	fmt.Print("\033[2J\033[H")
	fmt.Printf("Running%v\r\n", counter)
	fmt.Printf("H: %v\r\n", height)
	lineCount := len(lines)
	for i := range lineCount {
		fmt.Printf("%v\r\n", strings.TrimSuffix(string(lines[i]), "\n"))
	}
	fmt.Printf("\033[%v;0H>", height-1)
	counter++
}

func listenStdin(channel chan byte, context context.Context) {
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-context.Done():
			return
		default:
			b, _ := reader.ReadByte()
			channel <- b
		}
	}
}

func main() {
	descriptor := int(os.Stdin.Fd())
	prevState, err := term.MakeRaw(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	input := make(chan byte)
	wrapup := func() {
		close(input)
		term.Restore(descriptor, prevState)
	}

	defer wrapup()
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGABRT)

	go func() {
		<-sigch
		wrapup()
		os.Exit(0)
	}()

	counter := 0
	currFlags, err := unix.IoctlGetTermios(descriptor, unix.TCGETS)
	if err != nil {
		log.Fatal(err)
	}
	currFlags.Iflag |= unix.ICRNL
	currFlags.Lflag |= (unix.ISIG | unix.ECHO | unix.ECHONL | unix.ICANON | unix.ECHONL | unix.ECHOE)
	// fyi there's a TCSETS as well that applies the setting differently
	if err := unix.IoctlSetTermios(descriptor, unix.TCSETSF, currFlags); err != nil {
		log.Fatal(err)
	}
	// stdinReader := bufio.NewReader(os.Stdin)
	// input := make([]byte, 4)
	_, height, err := term.GetSize(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	// go listenStdin(input, context.Background())
	inputLines := make([][]byte, 0)
	buffer := bufio.NewReader(os.Stdin)
	draw(counter, height, inputLines)
	for {
		// b := <-input
		newLine, err := buffer.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}

		inputLines = append(inputLines, newLine)
		draw(counter, height, inputLines)
		counter++
	}
	// os.Stdin.Fd()
	// cmd.Execute()
}
