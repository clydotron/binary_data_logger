package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
)

const (
	writeChanBufferSize   = 20
	defaultWriteThreshold = time.Duration(time.Millisecond * 500)
)

type BinaryLogger interface {
	Write(loggable BinaryLoggable) error

	Read(file *os.File, clazz any) (SimpleIterator, error)
}

// Binary Logger implementation:
// supports multi-threaded usage
// uses a buffered channel internally to queue up all writes (Write call does not block)
// Has a go routine dedicated to consuming the queued loggables,
// serializing them, then appending the binary data to the log file.
// Uses buffered readers and writers (4096 bytes) to minimize disk io operations.
// Writes will be flushed as needed (buffer is full), or every t milliseconds as needed
// Handles partial reads: serialized data can span more than one buffer block

type BinaryLoggerImpl struct {
	maxLogFileSize int64
	writeCh        chan BinaryLoggable
	writeThreshold time.Duration
}

func NewBinaryLogger(ctx context.Context,
	wg *sync.WaitGroup,
	fileName string,
	maxLogFileSize int64) BinaryLogger {

	logger := BinaryLoggerImpl{
		writeCh:        make(chan BinaryLoggable, writeChanBufferSize),
		writeThreshold: defaultWriteThreshold,
		maxLogFileSize: maxLogFileSize,
	}

	wg.Add(1)
	go logger.writeToFile(ctx, wg, fileName)
	return &logger
}

func (logger *BinaryLoggerImpl) openLogFile(fileName string, fileId int) {
	// attempt to open the file.
	// if the file exists, check its size
	// if full, create a new file with an incremented id

}

// TODO finish implementing
func (logger *BinaryLoggerImpl) writeLoggable(writer *bufio.Writer, loggable BinaryLoggable) int64 {

	bytes, err := loggable.toBytes()
	if err != nil {
		log.Fatalln("failed to serialize log info:", err)
	}
	// check to see if we have room in the file
	//if logger.maxLogFileSize < fileSize+int64(len(bytes))+4 {

	// now what?
	// flush the writer, close the file
	// create new file and new writer
	//}

	var bytesWritten int64 = 0

	// maintain data consistency: if the writer fills the current buffer
	// and then writes to the next, flush the second one right away
	bytesToWrite := len(bytes)
	flushNow := writer.Available() < bytesToWrite

	nBytes, err := writer.WriteRune(rune(bytesToWrite))
	if err != nil {
		log.Fatalln("failed to write data size. error:", err)
	}
	bytesWritten += int64(nBytes)

	nBytes, err = writer.Write(bytes)
	if err != nil {
		log.Fatalln("failed to write data. error:", err)
	}
	bytesWritten += int64(nBytes)
	if flushNow {
		writer.Flush()
	}

	return bytesWritten
}

func (logger *BinaryLoggerImpl) writeToFile(ctx context.Context, wg *sync.WaitGroup, fileName string) {
	// if any of the code called from this function cause a panic,
	// recover and log the error
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered: error:", r)
		}
	}()
	defer close(logger.writeCh)
	defer wg.Done()

	// TODO design solitop to switch files on the fly:
	// fill one file up, move to next
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open the logger file")
	}
	defer f.Close()

	// get the current file size:
	var fileSize int64 = 0
	fi, err := f.Stat()
	if err != nil {
		fileSize = fi.Size()
	}

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	// use a ticker to automatically flush the buffer every t milliseconds
	ticker := time.NewTicker(time.Second)
	lastFlushTS := time.Now()

	for {
		select {
		case <-ctx.Done():
			log.Println(">> ending file size:", fileSize)
			return

		case <-ticker.C:
			// if we have buffered data, check how long its been since the last flush
			// if more than the specified threshold, flush it
			if writer.Buffered() > 0 {
				tdelta := time.Since(lastFlushTS)
				if tdelta > logger.writeThreshold {
					if err = writer.Flush(); err != nil {
						log.Fatalln("writer.Flush failed. error:", err)
					}
					lastFlushTS = time.Now()
				}
			}

		case loggable := <-logger.writeCh:

			// TODO move this into separate function (this one is too big)
			//logger.writeLoggable(writer, loggable)

			bytes, err := loggable.toBytes()
			if err != nil {
				log.Fatalln("failed to serialize log info:", err)
			}
			// check to see if we have room in the file
			//if logger.maxLogFileSize < fileSize+int64(len(bytes))+4 {

			// now what?
			// flush the writer, close the file
			// create new file and new writer
			//}

			// maintain data consistency: if the writer fills the current buffer
			// and then writes to the next, flush the second one right away
			bytesToWrite := len(bytes)
			flushNow := writer.Available() < bytesToWrite

			nBytes, err := writer.WriteRune(rune(bytesToWrite))
			if err != nil {
				log.Fatalln("failed to write data size. error:", err)
			}
			fileSize += int64(nBytes)

			nBytes, err = writer.Write(bytes)
			if err != nil {
				log.Fatalln("failed to write data. error:", err)
			}
			fileSize += int64(nBytes)
			if flushNow {
				writer.Flush()
			}
		}
	}
}

func (logger *BinaryLoggerImpl) Write(loggable BinaryLoggable) error {

	// provide a default for the case where the channel is full, otherwise this could block
	// if the channel is full, something has gone awry with the logger,
	// up to the caller to decide what to do next (retry, give up, ...)
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
