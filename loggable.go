package main

import (
	"bytes"
	"encoding/binary"
)

type BinaryLoggable interface {

	// Serialize the fields of this object into a byte array.
	toBytes() ([]byte, error)

	// Deserialize the fields of this object from given byte array.
	fromBytes(rawBytes []byte) error
}

type LoggableImpl1 struct {
	Val1 int32
	Val2 int32
	Val3 float32
	// make endian-ness a parameter?
}

// TODO Add a second one with variable sized data

func (l *LoggableImpl1) toBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, l)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *LoggableImpl1) fromBytes(rawBytes []byte) error {
	// check the parameters:

	buf := bytes.NewBuffer(rawBytes)

	// put the raw bytes into a reader
	err := binary.Read(buf, binary.LittleEndian, l)
	if err != nil {
		return err
	}

	return nil
}
