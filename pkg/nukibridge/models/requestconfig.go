package models

import (
	"bytes"

	log "github.com/sirupsen/logrus"
)

type RequestConfig struct {
	Nonce [32]byte
}

func EncodeRequestConfig(r RequestConfig) ([]byte, error) {
	payload := new(bytes.Buffer)
	if _, err := payload.Write(r.Nonce[:]); err != nil {
		log.WithError(err).Errorln("Failed to encode request config")
	}
	return payload.Bytes(), nil
}
