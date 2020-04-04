package models

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/mapero/nuki-bridge/pkg/nukibridge/enums"
	log "github.com/sirupsen/logrus"
)

type LogEntry struct {
	Index     uint32        `json:"index"`
	Timestamp time.Time     `json:"timestamp"`
	AuthID    uint32        `json:"authId"`
	Name      string        `json:"name"`
	Type      enums.LogType `json:"type"`
	Details   interface{}   `json:"details"`
}

// type Logging
type LogEntryTypeLogging struct {
	Logging bool `json:"logging"`
}

// type LockAction, Calibration, Initialization Run
type LogEntryTypeLockAction struct {
	LockAction       enums.LockAction       `json:"lockAction"`
	Trigger          enums.Trigger          `json:"trigger"`
	Flags            uint8                  `json:"flags"`
	CompletionStatus enums.CompletionStatus `json:"completionStatus"`
}

// type KeypadAction
type LogEntryTypeKeypadAction struct {
	LockAction       enums.LockAction         `json:"lockAction"`
	Source           enums.KeypadActionSource `json:"source"`
	CompletionStatus enums.CompletionStatus   `json:"completionStatus"`
	CodeID           uint16                   `json:"codeId"`
}

// type DoorSensor
type LogEntryTypeDoorSensor struct {
	DoorSensor enums.DoorSensor `json:"doorSensor"`
}

func DecodeLogEntry(b []byte) (entry LogEntry, err error) {
	r := bytes.NewReader(b)
	var data struct {
		Index  uint32
		Year   uint16
		Month  byte
		Day    byte
		Hour   byte
		Minute byte
		Second byte
		AuthID uint32
		Name   [32]byte
		Type   uint8
	}
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		log.WithError(err).Errorln("Failed to decode log entry")
		return entry, nil
	}
	entry.Index = data.Index
	entry.Timestamp = time.Date(
		int(data.Year),
		time.Month(data.Month),
		int(data.Day),
		int(data.Hour),
		int(data.Minute),
		int(data.Second),
		0,
		time.UTC)
	entry.AuthID = data.AuthID
	entry.Name = string(bytes.Trim(data.Name[:], "\x00"))
	entry.Type = enums.LogType(data.Type)
	switch entry.Type {
	case enums.LogTypeLogging, enums.LogTypeDoorSensorLogging:
		logging, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		entry.Details = LogEntryTypeLogging{
			Logging: logging == 0x01,
		}
	case enums.LogTypeLockAction, enums.LogTypeCalibration, enums.LogTypeInitializationRun:
		action, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		trigger, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		flags, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		status, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		entry.Details = LogEntryTypeLockAction{
			LockAction:       enums.LockAction(action),
			Trigger:          enums.Trigger(trigger),
			Flags:            flags,
			CompletionStatus: enums.CompletionStatus(status),
		}
	case enums.LogTypeKeypadAction:
		action, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		source, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		status, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		entrydata := LogEntryTypeKeypadAction{
			LockAction:       enums.LockAction(action),
			Source:           enums.KeypadActionSource(source),
			CompletionStatus: enums.CompletionStatus(status),
		}
		if err := binary.Read(r, binary.LittleEndian, &entrydata.CodeID); err != nil {
			log.WithError(err).Errorln("Failed to decode log entry")
			return entry, err
		}
		entry.Details = entrydata
	case enums.LogTypeDoorSensor:
		door, err := r.ReadByte()
		if err != nil {
			return entry, err
		}
		entry.Details = LogEntryTypeDoorSensor{
			DoorSensor: enums.DoorSensor(door),
		}
	}

	return entry, nil
}
