package nukibridge

import (
	"bytes"
	"encoding/binary"
)

type AuthorizationIDResponse struct {
	Authenticator   [32]byte
	AuthorizationID uint32
	UUID            [16]byte
	Nonce           [32]byte
}

func NewAuthoritationIDResponse(b []byte) (*AuthorizationIDResponse, error) {
	a := &AuthorizationIDResponse{}
	buf := bytes.NewBuffer(b)

	_, err := buf.Read(a.Authenticator[:])
	if err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.LittleEndian, &a.AuthorizationID); err != nil {
		return nil, err
	}

	_, err = buf.Read(a.UUID[:])
	if err != nil {
		return nil, err
	}

	_, err = buf.Read(a.Nonce[:])
	if err != nil {
		return nil, err
	}

	return a, nil
}
