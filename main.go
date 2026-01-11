package main

import (
	"github.com/UltimateForm/tcprcon/cmd"
)

func main() {
	cmd.Execute()
	// logger.Setup(logger.LevelDebug)
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// app := fullterm.CreateApp()
	// errChan := make(chan error, 1)
	// appRun := func() {
	// 	errChan <- app.Run(ctx)
	// }

	// appWriter := func() {
	// 	counter := 0
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			fmt.Fprintf(app, "Counter: %v\n\r", counter)
	// 			counter++
	// 			time.Sleep(time.Duration(time.Second * 1))
	// 		}
	// 	}
	// }
	// go appRun()
	// go appWriter()

	// if err := <-errChan; err != nil {
	// 	cancel()
	// 	logger.Critical.Fatal(app)
	// }
}
