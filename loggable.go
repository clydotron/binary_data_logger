package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type BinaryLoggable interface {

	// Serialize the fields of this object into a byte array.
	toBytes() ([]byte, error)

	// Deserialize the fields of this object from given byte array.
	fromBytes(rawBytes []byte) error
}

type LoggableImpl struct {
	DeviceId int32
	ReportId int32
	Value    float32
}

func (l *LoggableImpl) String() string {
	return fmt.Sprintf("DeviceId: %d ReportId: %d Val3: %f", l.DeviceId, l.ReportId, l.Value)
}

func (l *LoggableImpl) toBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, l)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *LoggableImpl) fromBytes(rawBytes []byte) error {
	// check the parameters:

	buf := bytes.NewBuffer(rawBytes)

	// put the raw bytes into a reader
	err := binary.Read(buf, binary.LittleEndian, l)
	if err != nil {
		return err
	}

	return nil
}
