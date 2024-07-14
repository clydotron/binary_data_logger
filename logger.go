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

type SimpleLogger interface {
	Write(loggable BinaryLoggable) error
	Read(file *os.File, clazz any) (SimpleIterator, error)
}

type BinaryLogger[T any] interface {
	Write(loggable BinaryLoggable) error //T

	// void write(T loggable) throws IOException;
	// this one will be a little tricky
	//Iterator<T> read(File file, Class<T> clazz) throws IOException;
	// in Go: Exceptions are Panic/Recover
	// experiment
	Read(file *os.File, clazz T) (IteratorT[T], error)
}

type BinaryLoggerImpl[T any] struct {
	fileName string
}

type SimpleLoggerImpl struct {
	fileName       string
	writeCh        chan BinaryLoggable
	writeThreshold time.Duration
}

// func NewBinaryLogger[T any](fileName string) BinaryLogger[T] {
// 	return &BinaryLoggerImpl[T]{
// 		fileName: fileName,
// 	}
// }

func NewSimpleLogger(ctx context.Context, fileName string) SimpleLogger {
	logger := SimpleLoggerImpl{
		fileName:       fileName,
		writeCh:        make(chan BinaryLoggable, writeChanBufferSize),
		writeThreshold: defaultWriteThreshold,
	}
	go logger.writeToFile(ctx)
	return &logger
}

func (logger *SimpleLoggerImpl) writeToFile(ctx context.Context) {

	f, err := os.OpenFile(logger.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open the logger file")
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	ticker := time.NewTicker(time.Second)
	lastFlushTS := time.Now()

	//lastUpdate := time.Now()
	// set a timer to automatically flush the buffer if its been longer than x millis
	for {
		//fmt.Println("inside")
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			// if we have buffered data, check how long it has been since we wrote it
			// if more than the specified threshold, flush it!
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
		}
	}
}

func (logger *SimpleLoggerImpl) Write(loggable BinaryLoggable) error {

	// provide a default for the case where the channel is full,
	// otherwise this could block
	select {
	case logger.writeCh <- loggable:
	default:
		return errors.New("channel is full")
	}

	return nil
}

func (logger *SimpleLoggerImpl) Read(file *os.File, clazz any) (SimpleIterator, error) {

	// create a reader for the file:
	reader := bufio.NewReader(file)

	// check to see if there is anything in the file:
	_, err := reader.Peek(4)
	hasNext := err == nil

	iter := &SimpleIteratorImpl{
		reader:     reader,
		returnType: reflect.TypeOf(LoggableImpl1{}),
		hasNext:    hasNext}

	return iter, nil
}

// ------------------------------------------------------------------------------------------------

func (logger *BinaryLoggerImpl[T]) Write(loggable BinaryLoggable) error {

	f, err := os.OpenFile(logger.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		//log.Fatal(err)
		// TODO revisit - Panic/Recover?
		return err
	}
	defer f.Close()

	bytes, err := loggable.toBytes()
	if err != nil {
		return err
	}

	// this might be overkill for what we want/need
	writer := bufio.NewWriter(f)
	// probably need to encode the length, or some separator
	writer.Write(bytes)
	writer.Flush()
	return nil
}

// func (logger *BinaryLoggerImpl[T]) Read(file *os.File, clazz T) (IteratorT[T], error) {

// 	// check the file pointer

// 	iter := IteratorImplT[T]{}

// 	return &iter, nil
// }

// func (logger *LoggerX) ReadT(file *os.File, t reflect.Type) error {
// 	// check parameters:

// 	// eventually we will need to create an iterator using the type
// 	reader := bufio.NewReader(file)
// 	buf := make([]byte, 2048)

// 	nBytes, err := reader.Read(buf)
// 	if err == io.EOF {

// 		return nil
// 	}

// 	fmt.Printf("read: %d bytes\n", nBytes)

// 	newItem := reflect.New(t)

// 	//if reflect.TypeOf(n).Implements(i)
// 	if reflect.TypeOf(t).Implements(reflect.TypeOf(new(BinaryLoggable)).Elem()) {

// 	}

// 	if newItem.Type().Implements(reflect.TypeOf(new(BinaryLoggable)).Elem()) {
// 		fmt.Println("ImplementsIt!")
// 	} else {
// 		fmt.Println("NoWorkie!")
// 	}

// 	// create a new iterator
// 	iter := IteratorT[]

// 	nx := newItem.Interface().(BinaryLoggable)
// 	nx.fromBytes(buf)

// 	fmt.Println(nx)
// 	// newItem.fromBytes()
// 	return nil
// }
