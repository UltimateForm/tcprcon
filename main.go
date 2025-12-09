package main

import (
	"context"
	"fmt"
	"time"

	"github.com/UltimateForm/tcprcon/internal/fullterm"
	"github.com/UltimateForm/tcprcon/internal/logger"
)

func main() {
	logger.Setup(logger.LevelDebug)
	ctx, cancel := context.WithCancel(context.Background())
	app := fullterm.CreateApp()
	errChan := make(chan error)
	appRun := func() {
		err := app.Run(ctx)
		errChan <- err
	}

	appWriter := func() {
		counter := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Fprintf(app, "Counter: %v\n\r", counter)
				counter++
				time.Sleep(time.Duration(time.Second * 1))
			}
		}
	}
	go appRun()
	go appWriter()

	err := <-errChan
	if err != nil {
		cancel()
		logger.Critical.Fatal(app)
	}
}
