package nukibridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mapero/nuki-bridge/pkg/nukibridge/api"
	"github.com/mapero/nuki-bridge/pkg/nukibridge/enums"
	log "github.com/sirupsen/logrus"
)

type SseEvent struct {
	Event string
	Data  interface{}
}

type NukiBridgeService struct {
	bridge *bridge

	callbackNotifier  chan api.CallbackObject
	newCallbacks      chan string
	removingCallbacks chan int
	callbacks         map[int]string

	sseNotifier       chan SseEvent
	newSseClients     chan chan SseEvent
	closingSseClients chan chan SseEvent
	sseClients        map[chan SseEvent]bool
}

func NewBridgeService(bridge *bridge) *NukiBridgeService {
	s := &NukiBridgeService{
		bridge:            bridge,
		callbackNotifier:  make(chan api.CallbackObject, 1),
		newCallbacks:      make(chan string),
		removingCallbacks: make(chan int),
		callbacks:         make(map[int]string),
		sseNotifier:       make(chan SseEvent, 1),
		newSseClients:     make(chan chan SseEvent),
		closingSseClients: make(chan chan SseEvent),
		sseClients:        make(map[chan SseEvent]bool),
	}
	go s.listen()

	return s
}

func (s *NukiBridgeService) listen() {
	for {
		select {

		case url := <-s.newCallbacks:

			// A new client has connected.
			// Register their message channel
			i := 0
			for {
				if _, ok := s.callbacks[i]; !ok {
					break
				}
				i++
			}
			s.callbacks[i] = url
			log.WithField("count", len(s.callbacks)).Infoln("Callback added")

		case id := <-s.removingCallbacks:
			// A client has dettached and we want to
			// stop sending them messages.
			delete(s.callbacks, id)
			log.WithField("count", len(s.callbacks)).Infoln("Callback removed")

		case event := <-s.callbackNotifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			body, err := json.Marshal(event)
			if err != nil {
				log.WithError(err).Errorln("Failed to receive callback event")
				continue
			}
			for _, url := range s.callbacks {
				_, err := http.Post(url, "application/json", bytes.NewBuffer(body))
				if err != nil {
					log.WithError(err).Errorln("Failed to send callback event")
				}
			}

		case c := <-s.newSseClients:
			s.sseClients[c] = true
			log.WithField("count", len(s.sseClients)).Infoln("SSE client added")

		case c := <-s.closingSseClients:
			delete(s.sseClients, c)
			log.WithField("count", len(s.sseClients)).Infoln("SSE client removed")

		case event := <-s.sseNotifier:
			for ch, _ := range s.sseClients {
				ch <- event
			}
		}

	}
}

func (s *NukiBridgeService) ListGet() (interface{}, error) {
	locks := s.bridge.GetLocks()

	list := make([]api.NukiLock, 0)
	for key, lock := range locks {
		entry := api.NukiLock{
			NukiId: int32(key),
			Name:   lock.lastConfig.Name,
			LastKnownState: api.LastLockState{
				State:           int32(lock.lastState.LockState),
				BatteryCritical: lock.lastState.CriticalBatteryState,
				StateName:       lock.lastState.LockState.String(),
				Timestamp:       lock.lastState.CurrentTime.String(),
			},
		}
		list = append(list, entry)
	}
	return list, nil
}
func (s *NukiBridgeService) LockStateGet(nukiId string) (interface{}, error) {
	id, err := strconv.ParseUint(nukiId, 10, 32)
	if err != nil {
		return nil, err
	}
	lock, err := s.bridge.GetLock(uint(id))
	if err != nil {
		return nil, err
	}
	s.bridge.aquireDevice()
	defer s.bridge.releaseDevice()
	state, err := lock.RequestKeyturnerState()
	if err != nil {
		return nil, err
	}
	return &api.NukiLockState{
		State:           int32(state.LockState),
		BatteryCritical: state.CriticalBatteryState,
		StateName:       state.LockState.String(),
		Success:         true,
	}, nil
}

func (s *NukiBridgeService) LockActionGet(nukiId string, action string, noWait string) (interface{}, error) {
	id, err := strconv.ParseUint(nukiId, 10, 32)
	if err != nil {
		return nil, err
	}
	act, err := strconv.ParseUint(action, 10, 32)
	if err != nil {
		return nil, err
	}
	lock, err := s.bridge.GetLock(uint(id))
	if err != nil {
		return nil, err
	}
	s.bridge.aquireDevice()
	defer s.bridge.releaseDevice()
	_, err = lock.LockAction(enums.LockAction(act), "")
	if err != nil {
		return nil, err
	}
	return &api.SimpleResponse{
		Success: true,
	}, nil
}

func (s *NukiBridgeService) CallbackAddGet(url string) (interface{}, error) {
	s.newCallbacks <- url
	return &api.SimpleResponse{
		Success: true,
	}, nil
}

func (s *NukiBridgeService) CallbackListGet() (interface{}, error) {
	callbacks := api.Callbacks{}
	for id, url := range s.callbacks {
		callback := api.Callback{
			Id:  int32(id),
			Url: url,
		}
		callbacks.Callbacks = append(callbacks.Callbacks, callback)
	}
	return callbacks, nil
}

func (s *NukiBridgeService) CallbackRemoveGet(nukiId string) (interface{}, error) {
	id, err := strconv.ParseUint(nukiId, 10, 32)
	if err != nil {
		return nil, err
	}
	s.removingCallbacks <- int(id)
	return &api.SimpleResponse{
		Success: true,
	}, nil
}

func (s *NukiBridgeService) LocksIdCurrentStateGet(id string) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	lock, err := s.bridge.GetLock(uint(nukiId))
	if err != nil {
		return nil, err
	}
	s.bridge.aquireDevice()
	defer s.bridge.releaseDevice()
	res, err := lock.RequestKeyturnerState()
	if err != nil {
		log.WithError(err).Errorln("Failed to request keyturner state")
		return nil, err
	}
	return res, nil
}

func (s *NukiBridgeService) LocksGet() (interface{}, error) {
	locks := make([]api.Lock, 0)
	for id, l := range s.bridge.GetLocks() {
		nukiId := fmt.Sprint(id)
		lock := api.Lock{
			Address: &l.address,
			Id:      &nukiId,
			Name:    &l.lastConfig.Name,
		}
		locks = append(locks, lock)
	}
	return locks, nil
}

func (s *NukiBridgeService) LocksIdHistoryGet(id string, offset string, count string) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	off, err := strconv.ParseUint(offset, 10, 32)
	if err != nil {
		return nil, err
	}
	c, err := strconv.ParseUint(count, 10, 32)
	if err != nil {
		return nil, err
	}
	lock, err := s.bridge.GetLock(uint(nukiId))
	if err != nil {
		return nil, err
	}
	s.bridge.aquireDevice()
	defer s.bridge.releaseDevice()
	return lock.RequestLogEntries(uint32(off), uint16(c))
}

func (s *NukiBridgeService) LocksIdLastStateGet(id string) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	lock, err := s.bridge.GetLock(uint(nukiId))
	if err != nil {
		return nil, err
	}
	return lock.lastState, nil
}

func (s *NukiBridgeService) BridgeConfigGet() (interface{}, error) {
	pairingEnabled := new(bool)
	*pairingEnabled = s.bridge.IsPairingEnabled()
	return api.BridgeConfig{
		PairingEnabled: pairingEnabled,
	}, nil
}

func (s *NukiBridgeService) BridgeConfigPut(bridgeConfig api.BridgeConfig) (interface{}, error) {
	if *bridgeConfig.PairingEnabled == true {
		s.bridge.EnablePairing()
	} else {
		s.bridge.DisablePairing()
	}
	return nil, nil
}

func (s *NukiBridgeService) LocksIdConfigGet(id string) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	lock, err := s.bridge.GetLock(uint(nukiId))
	if err != nil {
		return nil, err
	}
	s.bridge.aquireDevice()
	defer s.bridge.releaseDevice()
	c, err := lock.RequestConfig()
	if err != nil {
		return nil, err
	}
	timezoneOffset := int32(c.TimezoneOffset.Minutes())
	advertisingMode := int32(c.AdvertisingMode)
	fobAction1 := int32(c.FobAction1)
	fobAction2 := int32(c.FobAction2)
	fobAction3 := int32(c.FobAction3)
	homekitstatus := int32(c.HomeKitStatus)
	ledbrightness := int32(c.LEDBrightness)
	timezoneId := int32(c.TimezoneID)
	config := api.LockConfig{
		NukiId:           &id,
		Name:             &c.Name,
		AdvertisingMode:  &advertisingMode,
		AutoUnlatch:      &c.AutoUnlatch,
		ButtonEnabled:    &c.ButtonEnabled,
		DstMode:          &c.DSTMode,
		FirmwareVersion:  &c.FirmwareVersion,
		FobAction1:       &fobAction1,
		FobAction2:       &fobAction2,
		FobAction3:       &fobAction3,
		HardwareRevision: &c.HardwareRevision,
		HasFob:           &c.HasFob,
		HasKeypad:        &c.HasKeypad,
		HomeKitStatus:    &homekitstatus,
		Latitute:         &c.Latitude,
		Longitute:        &c.Longitude,
		LedBrightness:    &ledbrightness,
		LedEnabled:       &c.LEDEnabled,
		PairingEnabled:   &c.PairingEnabled,
		SingleLock:       &c.SingleLock,
		TimezoneId:       &timezoneId,
		TimezoneOffset:   &timezoneOffset,
	}
	return config, nil
}

// LocksIdDelete - Update a linked lock
func (s *NukiBridgeService) LocksIdDelete(id string) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	delete(s.bridge.Locks, uint(nukiId))
	s.bridge.saveConfig()
	return nil, nil
}

// LocksIdGet - Returns a linked lock
func (s *NukiBridgeService) LocksIdGet(id string) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	l, err := s.bridge.GetLock(uint(nukiId))
	if err != nil {
		return nil, err
	}
	return api.Lock{
		Address: &l.address,
		Id:      &id,
		Name:    &l.lastConfig.Name,
	}, nil

}

// LocksIdPut - Update a linked lock
func (s *NukiBridgeService) LocksIdPut(id string, lock api.Lock) (interface{}, error) {
	nukiId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	l, err := s.bridge.GetLock(uint(nukiId))
	if err != nil {
		return nil, err
	}
	l.adminPIN = uint(*lock.Pin)
	s.bridge.saveConfig()
	return nil, nil
}

func (s *NukiBridgeService) EventsGet(w http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing.
	//
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan SseEvent)

	// Signal the broker that we have a new connection
	s.newSseClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		s.closingSseClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	// notify := rw.(http.CloseNotifier).CloseNotify()
	notify := r.Context().Done()

	go func() {
		<-notify
		s.closingSseClients <- messageChan
	}()

	for {

		// Write to the ResponseWriter
		// Server Sent Events compatible
		event := <-messageChan
		data, err := json.Marshal(event.Data)
		if err != nil {
			log.WithError(err).Warnln("Failed to json marshal sse event")
			continue
		}
		fmt.Fprintf(w, "event: %s\n", event.Event)
		fmt.Fprintf(w, "data: %s\n\n", data)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}
	return
}
