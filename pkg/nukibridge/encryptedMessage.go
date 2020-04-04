package nukibridge

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/howeyc/crc16"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

type encryptedMessage struct {
	Nonce                [24]byte
	AuthorizationID      uint32
	MessageLength        uint16
	InnerAuthorizationID uint32
	CommandID            Command
	Payload              []byte
	CRC                  uint16
}

func (l *lock) receiveEncrypted(ch chan []byte) (messages []encryptedMessage, err error) {
	resp, err := l.receive(ch)
	if err != nil {
		return
	}
	var sharedKey [32]byte
	var peersPublicKey [32]byte
	copy(peersPublicKey[:], l.peersPublicKey)
	box.Precompute(&sharedKey, &peersPublicKey, &l.bridgePrivateKey)
	var unencrypted []byte
	r := bytes.NewBuffer(resp)
	for r.Len() > 0 {
		msg := encryptedMessage{}
		if _, err := r.Read(msg.Nonce[:]); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &msg.AuthorizationID); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &msg.MessageLength); err != nil {
			return nil, err
		}
		encrypted := make([]byte, msg.MessageLength)
		if _, err := r.Read(encrypted); err != nil {
			return nil, err
		}

		decrypted, ok := secretbox.Open(unencrypted, encrypted, &msg.Nonce, &sharedKey)
		if !ok {
			return nil, errors.New("Decrypt failed")
		}
		decryptedBuffer := bytes.NewBuffer(decrypted)
		msg.Payload = make([]byte, msg.MessageLength-secretbox.Overhead-8)
		if err := binary.Read(decryptedBuffer, binary.LittleEndian, &msg.InnerAuthorizationID); err != nil {
			return nil, err
		}
		if err := binary.Read(decryptedBuffer, binary.LittleEndian, &msg.CommandID); err != nil {
			return nil, err
		}
		if _, err := decryptedBuffer.Read(msg.Payload); err != nil {
			return nil, err
		}
		if err := binary.Read(decryptedBuffer, binary.LittleEndian, &msg.CRC); err != nil {
			return nil, err
		}
		log.WithField("lock", l.address).WithField("message", fmt.Sprintf("%+v", msg)).Debugln("Decrypted message")
		messages = append(messages, msg)
	}
	return messages, nil
}

func (l *lock) writeEncryptedMessage(c *ble.Characteristic, cmd uint16, payload []byte) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, cmd); err != nil {
		log.WithError(err).Errorln("Failed to write encrypted message")
		return err
	}
	if _, err := buf.Write(payload); err != nil {
		log.WithError(err).Errorln("Failed to write encrypted message")
		return err
	}
	msg, err := l.encodeEncryptedMessage(buf.Bytes())
	if err != nil {
		log.WithError(err).Errorln("Failed to write encrypted message")
		return err
	}
	if err := l.WriteCmd(c, msg); err != nil {
		log.WithError(err).Errorln("Failed to write encrypted message")
		return err
	}
	return nil
}

func (l *lock) writeEncryptedCmdRequest(c *ble.Characteristic, cmd uint16) error {
	req := new(bytes.Buffer)
	if err := binary.Write(req, binary.LittleEndian, cmd); err != nil {
		log.WithError(err).Errorln("Failed to write encrypted cmd request")
		return err
	}
	if err := l.writeEncryptedMessage(c, uint16(CmdRequestData), req.Bytes()); err != nil {
		log.WithError(err).Errorln("Failed to write encrypted cmd request")
		return err
	}
	return nil
}

func (l *lock) encodeEncryptedMessage(payload []byte) ([]byte, error) {
	// unencrypted part
	body := new(bytes.Buffer)
	if err := binary.Write(body, binary.LittleEndian, l.authorizationID); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	if _, err := body.Write(payload); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	crc := crc16.ChecksumCCITTFalse(body.Bytes())
	if err := binary.Write(body, binary.LittleEndian, crc); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	// encrypted part
	var sharedKey [32]byte
	var peersPublicKey [32]byte
	copy(peersPublicKey[:], l.peersPublicKey)
	box.Precompute(&sharedKey, &peersPublicKey, &l.bridgePrivateKey)
	var box []byte
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	box = secretbox.Seal(box, body.Bytes(), &nonce, &sharedKey)
	msg := new(bytes.Buffer)
	if _, err := msg.Write(nonce[:]); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	if err := binary.Write(msg, binary.LittleEndian, l.authorizationID); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	var length uint16 = uint16(body.Len() + secretbox.Overhead)

	if err := binary.Write(msg, binary.LittleEndian, length); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	if _, err := msg.Write(box); err != nil {
		log.WithError(err).Errorln("Failed to encode encrypted message")
		return nil, err
	}
	return msg.Bytes(), nil
}
