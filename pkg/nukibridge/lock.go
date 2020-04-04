package nukibridge

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"

	"github.com/go-ble/ble"
	"github.com/mapero/nuki-bridge/pkg/nukibridge/enums"
	"github.com/mapero/nuki-bridge/pkg/nukibridge/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
)

var (
	KeyturnerPairingServiceUUID = ble.MustParse("a92ee100-5501-11e4-916c-0800200c9a66")
	KeyturnerServiceUUID        = ble.MustParse("a92ee200-5501-11e4-916c-0800200c9a66")

	KeyturnerPairingServiceCharacteristicUUID        = ble.MustParse("a92ee101-5501-11e4-916c-0800200c9a66")
	KeyturnerServiceGDIOCharacteristicUUID           = ble.MustParse("a92ee201-5501-11e4-916c-0800200c9a66")
	KeyturnerServiceUSDIOCharacteristicUUID          = ble.MustParse("a92ee202-5501-11e4-916c-0800200c9a66")
	NukiRequestDataCmd                        uint16 = 0x0001
	NukiPublicKeyReqCmd                       uint16 = 0x0003
)

type Lock interface {
	Authenticate(publickey [32]byte, privateKey [32]byte) error
	Init()
	WriteCmd(c *ble.Characteristic, b []byte) error
	SubscribeIndicate(c *ble.Characteristic) (chan []byte, error)
}

type lock struct {
	address         string
	authorizationID uint32
	adminPIN        uint

	lastConfig models.Config
	lastState  models.KeyturnerStates

	bridgePublicKey  [32]byte
	bridgePrivateKey [32]byte
	peersPublicKey   []byte

	client           ble.Client
	connected        bool
	cancelConnection context.CancelFunc

	keyturnerPairingGDIO   *ble.Characteristic
	keyturnerGDIO          *ble.Characteristic
	keyturnerUSDIO         *ble.Characteristic
	chKeyturnerPairingGDIO chan []byte
	chKeyturnerGDIO        chan []byte
	chKeyturnerUSDIO       chan []byte
}

func NewLock(address string, authorizationID uint32, publicKey []byte, adminPIN uint) *lock {
	return &lock{
		address:         address,
		authorizationID: authorizationID,
		peersPublicKey:  publicKey,
		adminPIN:        adminPIN,
	}
}

func (l *lock) Connect() error {
	log.WithField("lock", l.address).Infoln("Connecting ...")
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.Addr().String()) == strings.ToUpper(l.address)
	}
	ctx, cancel := context.WithCancel(context.Background())
	l.cancelConnection = cancel
	client, err := ble.Connect(ctx, filter)
	if err != nil {
		log.WithError(err).Errorln("Failed to connect")
		return err
	}
	l.connected = true
	l.client = client
	client.ExchangeMTU(150)

	go func() {
		<-client.Disconnected()
		log.WithField("lock", l.address).Infoln("Lock disconnected")
		l.connected = false
	}()
	l.discover()
	return nil
}

func (l *lock) Disconnect() {
	if l.cancelConnection != nil {
		l.cancelConnection()
		l.connected = false
	}
}

func (l *lock) Init(publickey [32]byte, privateKey [32]byte) {
	log.WithField("lock", l.address).Infoln("Initializing lock instance")
	l.bridgePrivateKey = privateKey
	l.bridgePublicKey = publickey
	l.Connect()
	l.RequestKeyturnerState()
	l.RequestConfig()
}

func (l *lock) discover() {
	log.WithField("lock", l.address).Infoln("Disconvering GATT services")
	// Discovery services
	ss, err := l.client.DiscoverServices([]ble.UUID{
		KeyturnerPairingServiceUUID,
		KeyturnerServiceUUID,
	})
	if err != nil {
		log.WithField("lock", l.address).WithError(err).Errorln("Failed to discover services")
		return
	}
	for _, s := range ss {
		log.WithField("lock", l.address).WithField("service", s.UUID.String).Debugln("Discovering GATT characteristics")

		cs, err := l.client.DiscoverCharacteristics([]ble.UUID{
			KeyturnerPairingServiceCharacteristicUUID,
			KeyturnerServiceGDIOCharacteristicUUID,
			KeyturnerServiceUSDIOCharacteristicUUID,
		}, s)
		if err != nil {
			log.WithField("lock", l.address).WithField("service", s.UUID.String).WithError(err).Errorln("Failed to discover characteristics")
			continue
		}
		for _, c := range cs {
			log.WithField("lock", l.address).WithField("service", s.UUID.String).WithField("characteristic", c.UUID.String()).Debugln("Discovering GATT characteristics")

			_, err := l.client.DiscoverDescriptors(nil, c)
			if err != nil {
				log.WithField("lock", l.address).WithField("service", s.UUID.String).WithField("characteristic", c.UUID.String()).WithError(err).Errorln("Failed to discover GATT descriptors")
				continue
			}

			if c.UUID.Equal(KeyturnerPairingServiceCharacteristicUUID) {
				l.keyturnerPairingGDIO = c
				ch, err := l.SubscribeIndicate(l.keyturnerPairingGDIO)
				if err != nil {
					log.WithField("lock", l.address).WithField("service", s.UUID.String).WithField("characteristic", c.UUID.String()).WithError(err).Errorln("Failed to subscribe")
					return
				}
				l.chKeyturnerPairingGDIO = ch
			}
			if c.UUID.Equal(KeyturnerServiceGDIOCharacteristicUUID) {
				l.keyturnerGDIO = c
				ch, err := l.SubscribeIndicate(l.keyturnerGDIO)
				if err != nil {
					log.WithField("lock", l.address).WithField("service", s.UUID.String).WithField("characteristic", c.UUID.String()).WithError(err).Errorln("Failed to subscribe")
					return
				}
				l.chKeyturnerGDIO = ch
			}
			if c.UUID.Equal(KeyturnerServiceUSDIOCharacteristicUUID) {
				l.keyturnerUSDIO = c
				ch, err := l.SubscribeIndicate(l.keyturnerUSDIO)
				if err != nil {
					log.WithField("lock", l.address).WithField("service", s.UUID.String).WithField("characteristic", c.UUID.String()).WithError(err).Errorln("Failed to subscribe")
					return
				}
				l.chKeyturnerUSDIO = ch

			}
		}
	}
}

func (l *lock) Authenticate(publickey [32]byte, privateKey [32]byte) error {
	if !l.connected {
		if err := l.Connect(); err != nil {
			return err
		}
	}
	l.bridgePrivateKey = privateKey
	l.bridgePublicKey = publickey

	log.WithField("lock", l.address).Infoln("Authenticating")

	key, err := l.RequestPublicKey()
	if err != nil {
		log.WithError(err).Errorln("Failed to authenticate bridge")
		return err
	}
	copy(l.peersPublicKey[:], key[:32])

	challenge, err := l.SendPublicKey()
	if err != nil {
		log.WithError(err).Errorln("Failed to authenticate bridge")
		return err
	}

	challenge, err = l.SendAuthorizationAuthenticator(challenge)
	if err != nil {
		log.WithError(err).Errorln("Failed to authenticate bridge")
		return err
	}

	resp, err := l.SendAuthorizationData(0x01, 50, "GoBridge", challenge)
	if err != nil {
		log.WithError(err).Errorln("Failed to authenticate bridge")
		return err
	}

	err = l.SendAuthorizationIDConfirmation(resp.AuthorizationID, resp.Nonce)
	if err != nil {
		log.WithError(err).Errorln("Failed to authenticate bridge")
		return err
	}
	return nil
}

func (l *lock) createAuthenticator(r []byte) (authenticator [32]byte, err error) {
	var sharedKey [32]byte
	var peersPublicKey [32]byte
	copy(peersPublicKey[:], l.peersPublicKey)
	box.Precompute(&sharedKey, &peersPublicKey, &l.bridgePrivateKey)
	mac := hmac.New(sha256.New, sharedKey[:])
	if _, err := mac.Write(r); err != nil {
		return authenticator, err
	}
	hash := mac.Sum(nil)
	copy(authenticator[:], hash[:32])
	return authenticator, nil
}

func (l *lock) receive(ch chan []byte) ([]byte, error) {
	log.WithField("lock", l.address).Debugln("Waiting for response")
	var received = make([]byte, 0)
	timer := time.NewTimer(500 * time.Millisecond)
	for {
		select {
		case raw := <-ch:
			received = append(received, raw...)
			timer.Reset(500 * time.Millisecond)
		case <-timer.C:
			if len(received) <= 0 {
				return nil, errors.New("Timeout")
			}
			return received, nil
		}
	}

}

func (l *lock) decodeUnencrypted(received []byte) (*PDATA, error) {
	d, err := Decode(received)
	if err != nil {
		return nil, err
	}
	if d.Command == CmdErrorReport {
		var code uint8
		var cmd uint16
		buf := bytes.NewBuffer(d.Payload)
		binary.Read(buf, binary.LittleEndian, &code)
		binary.Read(buf, binary.LittleEndian, &cmd)
		return nil, fmt.Errorf("Lock reported error %x for command %x", code, cmd)
	}
	return d, nil
}

func (l *lock) WriteCmd(c *ble.Characteristic, b []byte) error {
	if err := l.client.WriteCharacteristic(c, b, false); err != nil {
		log.WithError(err).Errorln("Failed to write to lock")
		return err
	}
	return nil
}

func (l *lock) SubscribeIndicate(c *ble.Characteristic) (chan []byte, error) {
	ch := make(chan []byte)
	f := func(b []byte) {
		log.WithField("data", fmt.Sprintf("0x%x", b)).WithField("characteristic", c.UUID.String()).Debugln("Received bytes")
		ch <- b
	}
	log.WithField("characteristic", c.UUID.String()).Debugln("Subscribing GATT characteristic")
	if err := l.client.Subscribe(c, true, f); err != nil {
		return nil, err
	}
	return ch, nil
}

func (l *lock) RequestPublicKey() ([]byte, error) {
	log.WithField("lock", l.address).Infoln("Requesting public key")
	req := PDATA{
		Command: CmdRequestData,
		Payload: make([]byte, 2),
	}
	binary.LittleEndian.PutUint16(req.Payload, uint16(CmdPublicKey))
	if err := l.WriteCmd(l.keyturnerPairingGDIO, req.Encode()); err != nil {
		log.WithError(err).Errorln("Failed to request public key")
		return nil, err
	}
	received, err := l.receive(l.chKeyturnerPairingGDIO)
	peersPublicKeyResp, err := l.decodeUnencrypted(received)
	if err != nil {
		log.WithError(err).Errorln("Failed to request public key")
		return nil, err
	}
	if peersPublicKeyResp.Command != CmdPublicKey {
		err = errors.New("Received wrong response")
		log.WithError(err).WithField("response", peersPublicKeyResp.Command).Errorln("Failed to request public key")
		return nil, err
	}
	log.WithField("lock", l.address).Infoln("Received public key from lock")
	return peersPublicKeyResp.Payload, nil
}

func (l *lock) SendPublicKey() ([]byte, error) {
	log.WithField("lock", l.address).Infoln("Sending public key")
	req := PDATA{
		Command: CmdPublicKey,
		Payload: l.bridgePublicKey[:],
	}
	if err := l.WriteCmd(l.keyturnerPairingGDIO, req.Encode()); err != nil {
		log.WithError(err).Errorln("Failed to send public key")
		return nil, err
	}
	received, err := l.receive(l.chKeyturnerPairingGDIO)
	challengeResp, err := l.decodeUnencrypted(received)
	if err != nil {
		log.WithError(err).Errorln("Failed to send public key")
		return nil, err
	}
	if challengeResp.Command != CmdChallenge {
		err = errors.New("Received wrong response")
		log.WithError(err).WithField("response", challengeResp.Command).Errorln("Failed to send authorization authenticator")
		return nil, err
	}
	log.WithField("lock", l.address).Infoln("Received challenge")
	return challengeResp.Payload, nil
}

func (l *lock) SendAuthorizationAuthenticator(nonce []byte) ([]byte, error) {
	log.WithField("lock", l.address).Infoln("Send authorization authenticator")
	r := append(l.bridgePublicKey[:], l.peersPublicKey[:]...)
	r = append(r, nonce[:32]...)
	authenticator, err := l.createAuthenticator(r)
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization authenticator")
		return nil, err
	}
	req := PDATA{
		Command: CmdAuthorizationAuthenticator,
		Payload: authenticator[:],
	}
	if err := l.WriteCmd(l.keyturnerPairingGDIO, req.Encode()); err != nil {
		log.WithError(err).Errorln("Failed to send authorization authenticator")
		return nil, err
	}
	received, err := l.receive(l.chKeyturnerPairingGDIO)
	challengeResp, err := l.decodeUnencrypted(received)
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization authenticator")
		return nil, err
	}
	if challengeResp.Command != CmdChallenge {
		err = errors.New("Received wrong response")
		log.WithError(err).WithField("response", challengeResp.Command).Errorln("Failed to send authorization authenticator")
		return nil, err
	}
	log.WithField("lock", l.address).Infoln("Received challenge")
	return challengeResp.Payload, nil
}

func (l *lock) SendAuthorizationData(idType byte, bridgeID uint32, name string, challenge []byte) (*AuthorizationIDResponse, error) {
	bodyBuf := new(bytes.Buffer)
	if err := binary.Write(bodyBuf, binary.LittleEndian, idType); err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	if err := binary.Write(bodyBuf, binary.LittleEndian, bridgeID); err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	_, err := bodyBuf.Write([]byte(name)[:32])
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	var nonce [32]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	_, err = bodyBuf.Write(nonce[:])
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	body := bodyBuf.Bytes()
	r := append(body, challenge...)
	authenticator, err := l.createAuthenticator(r)
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	payload := new(bytes.Buffer)
	_, err = payload.Write(authenticator[:])
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	_, err = payload.Write(body)
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}

	req := PDATA{
		Command: CmdAuthorizationData,
		Payload: payload.Bytes(),
	}

	if err := l.WriteCmd(l.keyturnerPairingGDIO, req.Encode()); err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	received, err := l.receive(l.chKeyturnerPairingGDIO)
	pData, err := l.decodeUnencrypted(received)
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	if pData.Command != CmdAuthorizationID {
		err = errors.New("Received wrong response")
		log.WithError(err).WithField("command", pData.Command).Errorln("Failed to send authorization request")
		return nil, err
	}

	resp, err := NewAuthoritationIDResponse(pData.Payload)
	if err != nil {
		log.WithError(err).Errorln("Failed to send authorization request")
		return nil, err
	}
	l.authorizationID = resp.AuthorizationID

	return resp, nil
}

func (l *lock) SendAuthorizationIDConfirmation(authorizationID uint32, challenge [32]byte) error {
	log.WithField("lock", l.address).Infoln("Sending Authorization ID confirmation")
	r := new(bytes.Buffer)
	if err := binary.Write(r, binary.LittleEndian, authorizationID); err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}
	_, err := r.Write(challenge[:])
	if err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}
	authenticator, err := l.createAuthenticator(r.Bytes())
	if err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}
	payload := new(bytes.Buffer)
	_, err = payload.Write(authenticator[:])
	if err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}
	if err := binary.Write(payload, binary.LittleEndian, authorizationID); err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}

	req := PDATA{
		Command: CmdAuthorizationIDConfirmation,
		Payload: payload.Bytes(),
	}
	if err := l.WriteCmd(l.keyturnerPairingGDIO, req.Encode()); err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}
	received, err := l.receive(l.chKeyturnerPairingGDIO)
	_, err = l.decodeUnencrypted(received)
	if err != nil {
		log.WithError(err).Errorln("Failed to send Authrorization id confirmation")
		return err
	}
	log.WithField("lock", l.address).Infoln("Authorization id confirmed")
	return nil
}

func (l *lock) RequestKeyturnerState() (state models.KeyturnerStates, err error) {
	if !l.connected {
		if err := l.Connect(); err != nil {
			return state, err
		}
	}
	log.WithField("lock", l.address).Infoln("Request keyturner state")

	if err := l.writeEncryptedCmdRequest(l.keyturnerUSDIO, uint16(CmdKeyturnerStates)); err != nil {
		return state, err
	}
	messages, err := l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return state, err
	}
	if messages[0].CommandID != CmdKeyturnerStates {
		err := errors.New("Received wrong command")
		log.WithError(err).WithField("expected", CmdKeyturnerStates).WithField("actual", messages[0].CommandID).Errorln("Failed to request keyturner state")
		return state, err
	}
	state, err = models.DecodeKeyturnerStates(messages[0].Payload)
	if err != nil {
		return state, err
	}
	l.lastState = state
	return
}

func (l *lock) RequestLogEntries(offset uint32, count uint16) (entries []models.LogEntry, err error) {
	if !l.connected {
		if err := l.Connect(); err != nil {
			return entries, err
		}
	}
	log.WithField("lock", l.address).WithField("offset", offset).WithField("count", count).Infoln("Request log entries")

	if err := l.writeEncryptedCmdRequest(l.keyturnerUSDIO, uint16(CmdChallenge)); err != nil {
		return entries, err
	}
	payload, err := l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return entries, err
	}

	req := models.RequestLogEntries{
		StartIndex: offset,
		Count:      count,
		PIN:        uint16(l.adminPIN),
		SortOrder:  enums.SortOrderDecending,
		TotalCount: false,
	}
	copy(req.Nonce[:], payload[0].Payload)
	encoded, err := models.EncodeRequestLogEntries(req)
	if err != nil {
		return nil, err
	}
	if err := l.writeEncryptedMessage(l.keyturnerUSDIO, uint16(CmdRequestLogEntries), encoded); err != nil {
		return entries, err
	}
	messages, err := l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return entries, err
	}
	if messages[0].CommandID != CmdLogEntry {
		err := errors.New("Received wrong command")
		log.WithError(err).WithField("expected", CmdKeyturnerStates).WithField("actual", messages[0].CommandID).Errorln("Failed to request log entries")
		return entries, err
	}
	for _, message := range messages {
		if message.CommandID == CmdLogEntry {
			entry, err := models.DecodeLogEntry(message.Payload)
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (l *lock) LockAction(action enums.LockAction, description string) (b []byte, err error) {
	if !l.connected {
		if err := l.Connect(); err != nil {
			return b, err
		}
	}
	log.WithField("lock", l.address).WithField("action", action.String()).Infoln("Lock Action triggered")

	if err := l.writeEncryptedCmdRequest(l.keyturnerUSDIO, uint16(CmdChallenge)); err != nil {
		return b, err
	}
	messages, err := l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return b, err
	}
	if messages[0].CommandID != CmdChallenge {
		err := errors.New("Received wrong command")
		log.WithError(err).WithField("expected", CmdChallenge).WithField("actual", messages[0].CommandID).Errorln("Failed lock action")
		return nil, err
	}

	req := models.RequestLockAction{
		LockAction: action,
		AppID:      50,
		Flags:      0,
	}
	copy(req.NameSuffix[:], description)
	copy(req.Nonce[:], messages[0].Payload)
	encoded, err := models.EncodeRequestLockAction(req)
	if err != nil {
		return nil, err
	}
	if err := l.writeEncryptedMessage(l.keyturnerUSDIO, uint16(CmdLockAction), encoded); err != nil {
		return nil, err
	}
	messages, err = l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (l *lock) RequestConfig() (config models.Config, err error) {
	if !l.connected {
		if err := l.Connect(); err != nil {
			return config, err
		}
	}
	log.WithField("lock", l.address).Infoln("Request config")

	if err := l.writeEncryptedCmdRequest(l.keyturnerUSDIO, uint16(CmdChallenge)); err != nil {
		return config, err
	}
	messages, err := l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return config, err
	}
	if messages[0].CommandID != CmdChallenge {
		err := errors.New("Received wrong command")
		log.WithError(err).WithField("expected", CmdChallenge).WithField("actual", messages[0].CommandID).Errorln("Failed to request config")
		return config, err
	}
	req := models.RequestConfig{}
	copy(req.Nonce[:], messages[0].Payload)
	encoded, err := models.EncodeRequestConfig(req)
	if err != nil {
		return config, err
	}
	if err := l.writeEncryptedMessage(l.keyturnerUSDIO, uint16(CmdRequestConfig), encoded); err != nil {
		return config, err
	}
	messages, err = l.receiveEncrypted(l.chKeyturnerUSDIO)
	if err != nil {
		return config, err
	}
	if messages[0].CommandID != CmdConfig {
		err := errors.New("Received wrong command")
		log.WithError(err).WithField("expected", CmdConfig).WithField("actual", messages[0].CommandID).Errorln("Failed to request config")
		return config, err
	}
	config, err = models.DecodeConfig(messages[0].Payload)
	l.lastConfig = config
	return
}
