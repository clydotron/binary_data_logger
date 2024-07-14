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

type loggableImpl struct {
	Name     string
	DeviceId int32
	ReportId int32
	Value    float32
	Data     []byte
	order    binary.ByteOrder
}

func NewLoggable(name string, deviceId, reportId int32, value float32, data []byte) BinaryLoggable {
	return &loggableImpl{
		DeviceId: deviceId,
		ReportId: reportId,
		Value:    value,
		Data:     data,
		order:    binary.LittleEndian,
	}
}

func (l *loggableImpl) String() string {
	return fmt.Sprintf("Name: %s -- DeviceId: %d ReportId: %d Value: %f bytes:[%d]", l.Name, l.DeviceId, l.ReportId, l.Value, len(l.Data))
}

type temploggableImpl struct {
	DeviceId int32
	ReportId int32
	Value    float32
}

func (l *loggableImpl) writeBytes(data []byte, buf *bytes.Buffer) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(data))); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return err
	}
	return nil
}

func (l *loggableImpl) readBytes(buf *bytes.Buffer) ([]byte, error) {
	var nBytes int32
	err := binary.Read(buf, binary.LittleEndian, &nBytes)
	if err != nil {
		return nil, err
	}
	dataBuf := make([]byte, nBytes)
	err = binary.Read(buf, binary.LittleEndian, &dataBuf)
	if err != nil {
		return nil, err
	}
	return dataBuf, nil
}

// TODO use the temp / local struct to serialize in/out with
func (l *loggableImpl) toBytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	tempV := temploggableImpl{
		DeviceId: l.DeviceId,
		ReportId: l.ReportId,
		Value:    l.Value,
	}

	err := binary.Write(buf, binary.LittleEndian, tempV)
	if err != nil {
		return nil, err
	}

	if err = l.writeBytes(l.Data, buf); err != nil {
		return nil, err
	}

	if err = l.writeBytes([]byte(l.Name), buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (l *loggableImpl) fromBytes(rawBytes []byte) error {

	tempStruct := temploggableImpl{}

	buf := bytes.NewBuffer(rawBytes)
	err := binary.Read(buf, binary.LittleEndian, &tempStruct)
	if err != nil {
		return err
	}

	data, err := l.readBytes(buf)
	if err != nil {
		return err
	}
	name, err := l.readBytes(buf)
	if err != nil {
		return err
	}

	l.DeviceId = tempStruct.DeviceId
	l.ReportId = tempStruct.ReportId
	l.Value = tempStruct.Value
	l.Data = data
	l.Name = string(name)

	return nil
}
