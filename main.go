package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	logFileName = "test.log"
)

type App struct {
	logger    SimpleLogger
	doneCh    chan bool
	waitGroup *sync.WaitGroup
}

// readFromFile - uses the BinaryLogger Read interface to retrieve an interator to the logs

func (app *App) readFromFile(ctx context.Context, fileName string) {
	app.waitGroup.Add(1)
	defer app.waitGroup.Done()

	fs, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	iterator, err := app.logger.Read(fs, LoggableImpl{})
	if err != nil {
		log.Fatal(err)
	}

	// TODO remove this - add metion of 'tail'
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// sit in this loop until the app is exited.
	// check for new messages every second.
	// TODO revise this - tail wont work
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done!")
			return

		case <-ticker.C:
			for iterator.HasNext() {
				loggedData := iterator.Next()
				log, ok := loggedData.(*LoggableImpl)
				if ok {
					fmt.Println(log.String())
				}
			}
		}
	}
}

// logData - simple function to generate log entries at a specified frequency
func (app *App) logData(ctx context.Context, prefix string, frequencyInMS, runLengthInSec int) {

	app.waitGroup.Add(1)
	defer app.waitGroup.Done()

	ticker := time.NewTicker(time.Duration(frequencyInMS) * time.Millisecond)
	defer ticker.Stop()

	timer := time.NewTimer(time.Duration(runLengthInSec) * time.Second)
	defer timer.Stop()
	counter := 1

	for {
		select {
		case <-ctx.Done():
			log.Println(prefix, "terminated")
			return
		case <-timer.C:
			log.Println(prefix, "timed out")
			return

		case <-ticker.C:
			loggable := &LoggableImpl{
				DeviceId: int32(counter),
				ReportId: int32(counter * 2),
				Value:    float32(counter * 1.0),
			}
			err := app.logger.Write(loggable)
			if err != nil {
				fmt.Println(prefix, "Write error:", err)
				return
			}
			counter += 1
		}
	}
}

func main() {

	// start with a fresh log file each pass:
	err := os.Remove(logFileName)
	if err != nil {
		log.Fatalln("Failed to delete log file:", err)
	}

	ctx, cancelFcn := context.WithCancel(context.Background())

	app := App{
		waitGroup: &sync.WaitGroup{},
		logger:    NewSimpleLogger(ctx, logFileName),
		doneCh:    make(chan bool),
	}

	// setup a channel to be notified when ctrl-c is pressed
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// generate logs for t seconds, then cancel them and wait for the funcs to exit
	runLength := 3
	go app.logData(ctx, "process1", 100, runLength)
	go app.logData(ctx, "process2", 250, runLength)
	go app.logData(ctx, "process3", 150, runLength)

	app.waitGroup.Wait()
	log.Println("writers are done.")

	// read all of the log files from logger
	go app.readFromFile(ctx, logFileName)

	// wait for control-c to exit
	<-signalCh
	cancelFcn()

	fmt.Println("waiting for everything to shutdown")
	app.waitGroup.Wait()

	fmt.Println("all done")
}
