package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	NukiID           uint32
	Name             string
	Latitude         float32
	Longitude        float32
	AutoUnlatch      bool
	PairingEnabled   bool
	ButtonEnabled    bool
	LEDEnabled       bool
	LEDBrightness    uint8
	CurrentTime      time.Time
	TimezoneOffset   time.Duration
	DSTMode          bool
	HasFob           bool
	FobAction1       uint8
	FobAction2       uint8
	FobAction3       uint8
	SingleLock       bool
	AdvertisingMode  uint8
	HasKeypad        bool
	FirmwareVersion  string
	HardwareRevision string
	HomeKitStatus    uint8
	TimezoneID       uint16
}

func DecodeConfig(b []byte) (config Config, err error) {
	r := bytes.NewReader(b)
	var data struct {
		NukiID           uint32
		Name             [32]byte
		Latitude         float32
		Longitude        float32
		AutoUnlatch      byte
		PairingEnabled   byte
		ButtonEnabled    byte
		LEDEnabled       byte
		LEDBrightness    byte
		Year             uint16
		Month            byte
		Day              byte
		Hour             byte
		Minute           byte
		Second           byte
		TimezoneOffset   int16
		DSTMode          byte
		HasFob           byte
		FobAction1       byte
		FobAction2       byte
		FobAction3       byte
		SingleLock       byte
		AdvertisingMode  byte
		HasKeypad        byte
		FirmwareVersion  [3]byte
		HardwareRevision [2]byte
		HomeKitStatus    byte
		TimezoneID       uint16
	}

	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		log.WithError(err).Errorln("Failed to decode config")
		return config, nil
	}
	config.NukiID = data.NukiID
	config.Name = string(bytes.Trim(data.Name[:], "\x00"))
	config.Latitude = data.Latitude
	config.Longitude = data.Longitude
	config.AutoUnlatch = data.AutoUnlatch == 0x01
	config.PairingEnabled = data.PairingEnabled == 0x01
	config.ButtonEnabled = data.ButtonEnabled == 0x01
	config.LEDEnabled = data.LEDEnabled == 0x01
	config.LEDBrightness = data.LEDBrightness
	config.CurrentTime = time.Date(
		int(data.Year),
		time.Month(data.Month),
		int(data.Day),
		int(data.Hour),
		int(data.Minute),
		int(data.Second),
		0,
		time.UTC)
	config.TimezoneOffset = time.Duration(data.TimezoneOffset) * time.Minute
	config.DSTMode = data.DSTMode == 0x01
	config.HasFob = data.HasFob == 0x01
	config.FobAction1 = data.FobAction1
	config.FobAction2 = data.FobAction2
	config.FobAction3 = data.FobAction3
	config.SingleLock = data.SingleLock == 0x01
	config.AdvertisingMode = data.AdvertisingMode
	config.HasKeypad = data.HasKeypad == 0x01
	config.FirmwareVersion = fmt.Sprintf("%v.%v.%v", data.FirmwareVersion[0], data.FirmwareVersion[1], data.FirmwareVersion[2])
	config.HardwareRevision = fmt.Sprintf("%v.%v", data.HardwareRevision[0], data.HardwareRevision[1])
	config.HomeKitStatus = data.HomeKitStatus
	config.TimezoneID = data.TimezoneID
	return config, nil
}
