package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"reflect"
)

type SimpleIterator interface {
	HasNext() bool
	Next() interface{}
}

type SimpleIteratorImpl struct {
	reader     *bufio.Reader
	returnType reflect.Type
	hasNext    bool
}

func (si *SimpleIteratorImpl) HasNext() bool {
	return si.hasNext
}

func (si *SimpleIteratorImpl) Next() interface{} {

	// use si.returnType to create a new instance of the same type
	newItem := reflect.New(si.returnType)
	loggable := newItem.Interface().(BinaryLoggable)

	// Get the size of the log entry
	rval, _, err := si.reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			si.hasNext = false
			return nil
		} else {
			log.Fatalln("failed to read rune:", err)
		}
	}
	bytesToRead := int(rval)

	buf := make([]byte, bytesToRead)
	nBytes, err := si.reader.Read(buf)
	if err != nil {
		si.hasNext = false
		log.Fatalln("failed to read from reader:", err)
		return nil
	}

	// In the event that the current block doesnt have all the data,
	// repeat the read, which should trigger the reader to load the next block from disk
	// retry n times
	// TODO add retry logic
	if nBytes != bytesToRead {
		missingBytes := bytesToRead - nBytes
		//fmt.Println("didnt read enough bytes - missing:", missingBytes)
		tbytesRead, err := si.reader.Read(buf[nBytes:])
		if err != nil {
			log.Fatalln("failed to read from reader:", err)
		}
		if tbytesRead != missingBytes {
			log.Println("still missing bytes: read:", tbytesRead, "needed:", missingBytes)
		}
		// try again?

		return nil
	}

	err = loggable.fromBytes(buf)
	if err != nil {
		log.Fatalf("error while deserializing binary data: %v\n", err)
		return nil
	}

	// check if there is more data:
	_, err = si.reader.Peek(4)
	if err == io.EOF {
		log.Println("reached end of data.")
		si.hasNext = false
	}

	return loggable
}
