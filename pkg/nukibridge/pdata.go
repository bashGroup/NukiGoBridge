package nukibridge

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"

	"github.com/howeyc/crc16"
	log "github.com/sirupsen/logrus"
)

type PDATA struct {
	Command Command
	Payload []byte
}

func (d *PDATA) Encode() []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, d.Command); err != nil {
		log.WithError(err).Errorln("Failed to encode PDATA")
	}
	buf.Write(d.Payload)
	crc := crc16.ChecksumCCITTFalse(buf.Bytes())
	if err := binary.Write(buf, binary.LittleEndian, crc); err != nil {
		log.WithError(err).Errorln("Failed to encode PDATA")
	}
	return buf.Bytes()
}

func Decode(b []byte) (*PDATA, error) {
	d := &PDATA{
		Payload: make([]byte, 0),
	}
	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.LittleEndian, &d.Command); err != nil {
		log.WithError(err).Errorln("Failed to decode PDATA")
		return nil, err
	}
	body, err := ioutil.ReadAll(buf)
	if err != nil {
		log.WithError(err).Errorln("Failed to decode PDATA")
		return nil, err
	}
	payload := body[:len(body)-2]
	//crc := body[len(body)-2:]
	d.Payload = payload
	// ToDo: Check CRC
	return d, nil
}
