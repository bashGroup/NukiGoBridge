package models

import (
	"bytes"
	"encoding/binary"

	"github.com/mapero/nuki-bridge/pkg/nukibridge/enums"
	log "github.com/sirupsen/logrus"
)

type RequestLockAction struct {
	LockAction enums.LockAction
	AppID      uint32
	Flags      uint8
	NameSuffix [20]byte
	Nonce      [32]byte
}

func EncodeRequestLockAction(r RequestLockAction) ([]byte, error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, r.LockAction); err != nil {
		log.WithError(err).Errorln("Failed to encode lock action")
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, r.AppID); err != nil {
		log.WithError(err).Errorln("Failed to encode lock action")
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, r.Flags); err != nil {
		log.WithError(err).Errorln("Failed to encode lock action")
		return nil, err
	}
	if _, err := payload.Write(r.NameSuffix[:]); err != nil {
		log.WithError(err).Errorln("Failed to encode lock action")
		return nil, err
	}

	if _, err := payload.Write(r.Nonce[:]); err != nil {
		log.WithError(err).Errorln("Failed to encode lock action")
		return nil, err
	}
	return payload.Bytes(), nil
}
