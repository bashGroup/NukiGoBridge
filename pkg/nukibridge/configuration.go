package nukibridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"golang.org/x/crypto/nacl/box"
)

type Configuration struct {
	PrivateKey string                       `json:"privateKey"`
	PublicKey  string                       `json:"publicKey"`
	Locks      map[string]LockConfiguration `json:"locks"`
}

type LockConfiguration struct {
	PublicKey       string `json:"publicKey"`
	Address         string `json:"address"`
	AuthorizationId string `json:"authorizationId"`
	AdminPIN        uint   `json:"adminPIN"`
}

func (b *bridge) init() error {
	rand, err := os.Open("/dev/urandom")
	if err != nil {
		return err
	}
	defer func() {
		rand.Close()
	}()
	pub, priv, err := box.GenerateKey(rand)
	if err != nil {
		return err
	}
	b.PrivateKey = *priv
	b.PublicKey = *pub
	if err := b.saveConfig(); err != nil {
		return err
	}
	return nil
}

func (b *bridge) loadConfig() error {
	f, err := os.Open(path.Join(b.dir, filename))
	if err != nil {
		return err
	}
	defer func() {
		f.Sync()
		f.Close()
	}()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	var cfg Configuration
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return err
	}
	privateKey, err := base64.StdEncoding.DecodeString(cfg.PrivateKey)
	if err != nil {
		return err
	}
	copy(b.PrivateKey[:], privateKey)
	publicKey, err := base64.StdEncoding.DecodeString(cfg.PublicKey)
	if err != nil {
		return err
	}
	copy(b.PublicKey[:], publicKey)
	for id, lockCfg := range cfg.Locks {
		nukiId, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return err
		}
		authorizationID, err := strconv.ParseUint(lockCfg.AuthorizationId, 10, 32)
		if err != nil {
			return err
		}
		publicKey, err := base64.StdEncoding.DecodeString(lockCfg.PublicKey)
		if err != nil {
			return err
		}
		lock := NewLock(lockCfg.Address, uint32(authorizationID), publicKey, lockCfg.AdminPIN)
		b.Locks[uint(nukiId)] = lock
	}
	return nil
}

func (b *bridge) saveConfig() error {
	cfg := Configuration{
		PrivateKey: base64.StdEncoding.EncodeToString(b.PrivateKey[:]),
		PublicKey:  base64.StdEncoding.EncodeToString(b.PublicKey[:]),
		Locks:      make(map[string]LockConfiguration),
	}
	for key, lock := range b.Locks {
		lockCfg := LockConfiguration{
			Address:         lock.address,
			AuthorizationId: fmt.Sprint(lock.authorizationID),
			PublicKey:       base64.StdEncoding.EncodeToString(lock.peersPublicKey[:]),
			AdminPIN:        lock.adminPIN,
		}
		cfg.Locks[fmt.Sprint(key)] = lockCfg
	}
	f, err := os.OpenFile(path.Join(b.dir, filename), os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return err
	}
	defer func() {
		f.Sync()
		f.Close()
	}()
	bytes, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}
	if _, err := f.Write(bytes); err != nil {
		return err
	}
	return nil
}
