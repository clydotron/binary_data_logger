package main

import (
	"context"
	"crypto/rand"
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
	logger    BinaryLogger
	doneCh    chan bool
	waitGroup *sync.WaitGroup
}

// readFromFile - uses the BinaryLogger Read interface to retrieve an interator to the logs

func (app *App) readFromFile(ctx context.Context, fileName string) {
	log.Println("ReadFromFile")
	// TODO add recover here to handle any panics in the Read machinery

	defer app.waitGroup.Done()

	fs, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	iterator, err := app.logger.Read(fs, loggableImpl{})
	if err != nil {
		log.Fatal(err)
	}

	// sit in this loop until either the program is exited, or we consume all of the data.
	// TODO explore adding tail functionality
	for {
		select {
		case <-ctx.Done():
			log.Println("Done!")
			return

		default:
			if !iterator.HasNext() {
				app.doneCh <- true
				return
			}

			loggedData := iterator.Next()
			logImpl, ok := loggedData.(*loggableImpl)
			if ok {
				fmt.Println(logImpl.String())
			}
		}
	}
}

// logData - simple function to generate log entries at a specified frequency
func (app *App) logData(ctx context.Context, prefix string, frequencyInMS, runLengthInSec int) {
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
			data := make([]byte, 512)
			rand.Read(data)
			loggable := &loggableImpl{
				Name:     prefix,
				DeviceId: int32(counter),
				ReportId: int32(counter * 2),
				Value:    float32(counter * 1.0),
				Data:     data,
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

	// start with a fresh log file each pass: (dont care about the error)
	os.Remove(logFileName)

	ctx, cancelFcn := context.WithCancel(context.Background())

	app := App{
		waitGroup: &sync.WaitGroup{},
		logger:    NewBinaryLogger(ctx, logFileName),
		doneCh:    make(chan bool),
	}

	// setup a channel to be notified when ctrl-c is pressed
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// generate logs for t seconds, then wait for the go routines to exit before we start reading
	runLength := 5
	// increment the wait group here so we dont have a race condition with the go rountines starting up
	app.waitGroup.Add(3)
	go app.logData(ctx, "process1", 10, runLength)
	go app.logData(ctx, "process2", 25, runLength)
	go app.logData(ctx, "process3", 15, runLength)

	app.waitGroup.Wait()

	// read all of the log files from logger
	app.waitGroup.Add(1)
	go app.readFromFile(ctx, logFileName)

	// wait for either control-c, or all the log messages to be read
	select {
	case <-signalCh:
		log.Println("ctrl-c detected")
	case <-app.doneCh:
	}

	cancelFcn()

	fmt.Println("waiting for everything to shutdown")
	app.waitGroup.Wait()

	fmt.Println("all done")
}
