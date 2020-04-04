package models

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/mapero/nuki-bridge/pkg/nukibridge/enums"
	log "github.com/sirupsen/logrus"
)

type KeyturnerStates struct {
	NukiState                      enums.NukiState        `json:"nukiState"`
	LockState                      enums.LockState        `json:"lockState"`
	Trigger                        enums.Trigger          `json:"trigger"`
	CurrentTime                    time.Time              `json:"currentTime"`
	TimezoneOffset                 time.Duration          `json:"timezoneOffset"`
	CriticalBatteryState           bool                   `json:"criticalBatteryState"`
	ConfigUpdateCount              uint8                  `json:"configUpdateCount"`
	LocknGoTimer                   bool                   `json:"locknGoTimer"`
	LastLockAction                 enums.LockAction       `json:"lastLockAction"`
	LastLockActionTrigger          enums.Trigger          `json:"lastLockActionTrigger"`
	LastLockActionCompletionStatus enums.CompletionStatus `json:"lastLockActionCompletionStatus"`
	DoorSensorState                enums.DoorSensorState  `json:"doorSensorState"`
}

func DecodeKeyturnerStates(b []byte) (states KeyturnerStates, err error) {
	r := bytes.NewReader(b)
	var data struct {
		NukiState                      byte
		LockState                      byte
		Trigger                        byte
		Year                           uint16
		Month                          byte
		Day                            byte
		Hour                           byte
		Minute                         byte
		Second                         byte
		TimezoneOffset                 int16
		CriticalBatteryState           byte
		ConfigUpdateCount              byte
		LocknGoTimer                   byte
		LastLockAction                 byte
		LastLockActionTrigger          byte
		LastLockActionCompletionStatus byte
		DoorSensorState                byte
	}

	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		log.WithError(err).Errorln("Failed to decode config")
		return states, nil
	}
	states.NukiState = enums.NukiState(data.NukiState)
	states.LockState = enums.LockState(data.LockState)
	states.Trigger = enums.Trigger(data.Trigger)
	states.CurrentTime = time.Date(
		int(data.Year),
		time.Month(data.Month),
		int(data.Day),
		int(data.Hour),
		int(data.Minute),
		int(data.Second),
		0,
		time.UTC)
	states.TimezoneOffset = time.Duration(data.TimezoneOffset) * time.Minute
	states.CriticalBatteryState = data.CriticalBatteryState == 0x01
	states.ConfigUpdateCount = data.ConfigUpdateCount
	states.LocknGoTimer = data.LocknGoTimer == 0x01
	states.LastLockAction = enums.LockAction(data.LastLockAction)
	states.LastLockActionTrigger = enums.Trigger(data.LastLockActionTrigger)
	states.LastLockActionCompletionStatus = enums.CompletionStatus(data.LastLockActionCompletionStatus)
	states.DoorSensorState = enums.DoorSensorState(data.DoorSensorState)
	return states, nil
}
