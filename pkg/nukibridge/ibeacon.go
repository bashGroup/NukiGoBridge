package nukibridge

import (
	"bytes"
	"encoding/binary"
	"errors"

	log "github.com/sirupsen/logrus"
)

type IBeacon struct {
	Company [2]byte
	Type    byte
	Length  uint8
	UUID    [16]byte
	NukiID  uint32
	Dirty   bool
}

func decodeIBeacon(b []byte) (beacon IBeacon, err error) {
	r := bytes.NewReader(b)
	var data struct {
		Company [2]byte
		Type    byte
		Length  uint8
		UUID    [16]byte
		NukiID  uint32
		TxPower int8
	}

	if r.Len() != 25 {
		return beacon, errors.New("Beacon has wrong size")
	}
	if err := binary.Read(r, binary.BigEndian, &data); err != nil {
		return beacon, err
	}
	if data.Type != 0x02 {
		return beacon, errors.New("Not an iBeacon")
	}
	beacon.Company = data.Company
	beacon.Type = data.Type
	beacon.Length = data.Length
	beacon.UUID = data.UUID
	beacon.NukiID = data.NukiID
	log.WithField("TxPower", data.TxPower).Debugln("Decoding")
	beacon.Dirty = data.TxPower != -60
	return beacon, nil
}
