package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"reflect"
	"time"
)

const (
	writeChanBufferSize   = 20
	defaultWriteThreshold = time.Duration(time.Second)
)

type BinaryLogger interface {
	Write(loggable BinaryLoggable) error

	Read(file *os.File, clazz any) (SimpleIterator, error)
}

type BinaryLoggerImpl struct {
	writeCh        chan BinaryLoggable
	writeThreshold time.Duration
}

func NewBinaryLogger(ctx context.Context, fileName string) BinaryLogger {
	logger := BinaryLoggerImpl{
		writeCh:        make(chan BinaryLoggable, writeChanBufferSize),
		writeThreshold: defaultWriteThreshold,
	}
	go logger.writeToFile(ctx, fileName)
	return &logger
}

func (logger *BinaryLoggerImpl) writeToFile(ctx context.Context, fileName string) {
	// if any of the code called from this function cause a panic,
	// recover and log the error
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered: error:", r)
		}
	}()
	defer close(logger.writeCh)

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open the logger file")
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	// use a ticker to automatically flush the buffer every t milliseconds
	ticker := time.NewTicker(time.Second)
	lastFlushTS := time.Now()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			// if we have buffered data, check how long its been since the last flush
			// if more than the specified threshold, flush it
			if writer.Buffered() > 0 {
				tdelta := time.Since(lastFlushTS)
				if tdelta > logger.writeThreshold {
					writer.Flush()
					lastFlushTS = time.Now()
				}
			}

		case loggable := <-logger.writeCh:
			bytes, err := loggable.toBytes()
			if err != nil {
				log.Fatalf("failed to serialize log info: %v\n", err)
			}
			writer.WriteRune(rune(len(bytes)))
			writer.Write(bytes)
			log.Println("logging:", loggable)
		}
	}
}

func (logger *BinaryLoggerImpl) Write(loggable BinaryLoggable) error {

	// provide a default for the case where the channel is full,
	// otherwise this could block
	select {
	case logger.writeCh <- loggable:
	default:
		return errors.New("channel is full")
	}

	return nil
}

func (logger *BinaryLoggerImpl) Read(file *os.File, clazz any) (SimpleIterator, error) {
	reader := bufio.NewReader(file)

	// check to see if there is anything in the file:
	_, err := reader.Peek(4)
	hasNext := err == nil

	iter := &SimpleIteratorImpl{
		reader:     reader,
		returnType: reflect.TypeOf(clazz),
		hasNext:    hasNext}

	return iter, nil
}
