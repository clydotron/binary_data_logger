package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
)

type SimpleIterator interface {
	HasNext() bool
	Next() interface{}
}

// alternatives: interface{} or any
type IteratorI interface {
	HasNext() bool
	Next() interface{}
}

// or
type IteratorAny interface {
	HasNext() bool
	Next() any
}

// or Generic
type IteratorT[T any] interface {
	HasNext() bool
	GetNext() T
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

	// use si.source to create a new item
	newItem := reflect.New(si.returnType)
	loggable := newItem.Interface().(BinaryLoggable)

	rval, size, err := si.reader.ReadRune()
	if err != nil {
		// check if EOF:
		if errors.Is(err, io.EOF) {
			fmt.Println("reached end of data.")
			si.hasNext = false
			return nil
		} else {
			log.Fatalf("failed to read rune: %v\n", err)
		}
	}
	fmt.Printf("size %d >> rune:%v\n", size, rval)
	bytesToRead := int(rval)

	buf := make([]byte, bytesToRead)
	nBytes, err := si.reader.Read(buf)
	if err != nil {
		// do something drastic...
		si.hasNext = false
		return nil
	}

	if nBytes != bytesToRead {
		fmt.Println("didnt read enough bytes!")
		return nil
	}

	err = loggable.fromBytes(buf)
	if err != nil {
		fmt.Printf("error while deserializing binary data: %v\n", err)
		return nil
	}

	// check if there is more data:
	_, err = si.reader.Peek(4)
	if err == io.EOF {
		si.hasNext = false
	}

	return loggable
}

// type IteratorImplT[T any] struct {
// 	hasNext bool
// 	reader  io.Reader
// }

// func (ii *IteratorImplT[T]) HasNext() bool {
// 	return ii.hasNext
// }

// func (ii *IteratorImplT[T]) GetNext() T {

// 	// decide how we want read to work:
// 	// do we read blocks of data from file, then iterate over those
// 	// for instances of T, or do we read each T one at a time...

// 	err := ii.reader.Read()
// 	// create a new instance of T
// 	val := new(T)
// 	//convert to known type:

// 	return val
// }
