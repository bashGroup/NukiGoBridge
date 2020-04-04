package models

import (
	"bytes"
	"encoding/binary"

	"github.com/mapero/nuki-bridge/pkg/nukibridge/enums"
	log "github.com/sirupsen/logrus"
)

type RequestLogEntries struct {
	StartIndex uint32
	Count      uint16
	SortOrder  enums.SortOrder
	TotalCount bool
	Nonce      [32]byte
	PIN        uint16
}

func EncodeRequestLogEntries(r RequestLogEntries) ([]byte, error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, r.StartIndex); err != nil {
		log.WithError(err).Errorln("Failed to encode request log entries")
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, r.Count); err != nil {
		log.WithError(err).Errorln("Failed to encode request log entries")
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, r.SortOrder); err != nil {
		log.WithError(err).Errorln("Failed to encode request log entries")
		return nil, err
	}
	totalcount := uint8(0)
	if r.TotalCount {
		totalcount = 1
	}
	if err := binary.Write(payload, binary.LittleEndian, totalcount); err != nil {
		log.WithError(err).Errorln("Failed to encode request log entries")
		return nil, err
	}
	if _, err := payload.Write(r.Nonce[:]); err != nil {
		log.WithError(err).Errorln("Failed to encode request log entries")
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, r.PIN); err != nil {
		log.WithError(err).Errorln("Failed to encode request log entries")
		return nil, err
	}
	return payload.Bytes(), nil
}
