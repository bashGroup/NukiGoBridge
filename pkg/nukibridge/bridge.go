package nukibridge

//go:generate go run -tags=dev assets/generate.go

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/mapero/nuki-bridge/pkg/nukibridge/api"
	"github.com/mapero/nuki-bridge/pkg/nukibridge/assets/templates"
	"github.com/mapero/nuki-bridge/pkg/nukibridge/models"
)

var (
	filename = "bridge.json"
)

type Bridge interface {
	GetLocks() map[uint]*lock
	GetLock(id uint) (*lock, error)
}

type bridge struct {
	PublicKey      [32]byte       `json:"public_key"`
	PrivateKey     [32]byte       `json:"private_key"`
	Locks          map[uint]*lock `json:"locks"`
	dir            string
	service        *NukiBridgeService
	advCh          map[string]chan ble.Advertisement
	deviceLock     chan bool
	cancelScan     context.CancelFunc
	scanCtx        context.Context
	token          string
	port           string
	skipAdv        chan bool
	pairingEnabled bool
}

func (b *bridge) EnablePairing() {
	if b.pairingEnabled {
		return
	}
	b.pairingEnabled = true
	timer := time.NewTimer(10 * time.Second)
	go func() {
		<-timer.C
		b.DisablePairing()
	}()
	log.Infoln("Pairing mode enabled for 10 sec")

}

func (b *bridge) DisablePairing() {
	log.Infoln("Pairing mode disabled")
	b.pairingEnabled = false
}

func (b *bridge) IsPairingEnabled() bool {
	return b.pairingEnabled
}

func (b *bridge) GetLocks() map[uint]*lock {
	return b.Locks
}

func (b *bridge) GetLock(id uint) (*lock, error) {
	l, ok := b.Locks[id]
	if !ok {
		return nil, errors.New("Not found")
	}
	return l, nil
}

func NewBridge(dir string, port string, token string) (Bridge, error) {
	log.Println("Creating new bridge")

	b := &bridge{
		dir:        dir,
		Locks:      make(map[uint]*lock),
		deviceLock: make(chan bool, 1),
		token:      token,
		port:       port,
		skipAdv:    make(chan bool, 1),
	}
	if _, err := os.Stat(path.Join(dir, filename)); err != nil {
		if err := b.init(); err != nil {
			return nil, err
		}
	}
	if err := b.loadConfig(); err != nil {
		return nil, err
	}
	if len(b.PrivateKey) == 0 || len(b.PublicKey) == 0 {
		if err := b.init(); err != nil {
			return nil, err
		}
	}
	dev, err := linux.NewDevice()
	if err != nil {
		panic(err)
	}

	ble.SetDefaultDevice(dev)

	log.Println("Initializing known locks")
	for _, lock := range b.Locks {
		lock.Init(b.PublicKey, b.PrivateKey)
	}

	go b.startAPIService()

	b.startAdvertisingMonitor()

	return b, nil
}

func (b *bridge) aquireDevice() {
	log.Println("Aquiring device lock")
	b.deviceLock <- true
	b.cancelScan()
	<-b.scanCtx.Done()
	time.Sleep(500 * time.Millisecond)
	log.Println("Lock aquired")
}

func (b *bridge) releaseDevice() {
	log.Println("Release lock")
	<-b.deviceLock
	b.startAdvertisingMonitor()
}

func (b *bridge) addAndAuthorizeLock(address string) {
	lock := &lock{
		address: address,
	}
	if err := lock.Connect(); err != nil {
		log.WithField("lock", address).WithError(err).Errorln("Failed to add and authorize lock")
		return
	}
	if err := lock.Authenticate(b.PublicKey, b.PrivateKey); err != nil {
		log.WithField("lock", address).WithError(err).Errorln("Failed to add and authorize lock")
		return
	}
	config, err := lock.RequestConfig()
	if err != nil {
		log.WithField("lock", address).WithError(err).Errorln("Failed to add and authorize lock")
		return
	}
	lock.Disconnect()
	b.Locks[uint(config.NukiID)] = lock
	b.saveConfig()
}

func (b *bridge) startAdvertisingMonitor() {
	log.Infoln("Monitoring advertisment")
	filter := func(a ble.Advertisement) bool {
		return strings.HasPrefix(strings.ToUpper(a.Addr().String()), "54:D2:72:")
	}
	advHandler := func(a ble.Advertisement) {
		select {
		case b.skipAdv <- true:
			defer func() { <-b.skipAdv }()
			if b.IsPairingEnabled() && len(a.ServiceData()) > 0 && a.ServiceData()[0].UUID.String() == "a92ee100550111e4916c0800200c9a66" {
				address := strings.ToUpper(a.Addr().String())
				for _, lock := range b.Locks {
					if lock.address == address {
						return
					}
				}
				log.WithField("lock", address).Infoln("Adding and authorizing lock")
				b.aquireDevice()
				b.addAndAuthorizeLock(address)
				b.releaseDevice()
			}
			if len(a.ManufacturerData()) == 25 {
				beacon, err := decodeIBeacon(a.ManufacturerData())
				if err != nil {
					log.WithError(err).Debugln("Failed to parse iBeacon, ignoring")
					return
				}
				log.WithField("data", fmt.Sprintf("%+v", beacon)).Debugln("Received beacon advertismenent from nuki device")
				lock, err := b.GetLock(uint(beacon.NukiID))
				if err != nil {
					log.WithError(err).Debugln("Skipping")
					return
				}
				if !beacon.Dirty || time.Since(lock.lastState.CurrentTime).Seconds() < 2 {
					return
				}
				b.aquireDevice()
				lock.Connect()
				state, err := lock.RequestKeyturnerState()
				if err != nil {
					log.WithError(err).Errorln("Failed to update lock state due to error")
				}
				lock.Disconnect()
				b.releaseDevice()
				log.WithField("state", fmt.Sprintf("%+v", state)).WithField("nukiID", beacon.NukiID).Debugln("Received state")
				b.service.callbackNotifier <- api.CallbackObject{
					DeviceType:      0x02,
					BatteryCritical: state.CriticalBatteryState,
					Mode:            int32(state.NukiState),
					NukiId:          int32(beacon.NukiID),
					State:           int32(state.LockState),
					StateName:       state.LockState.String(),
				}
				data := struct {
					models.KeyturnerStates
					NukiId uint32
				}{
					state,
					beacon.NukiID,
				}
				b.service.sseNotifier <- SseEvent{
					Event: "state",
					Data:  data,
				}
			}
		default:
			log.Debugln("Skipping advertisment")
			return
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.scanCtx = ctx
	b.cancelScan = cancel
	go func() {
		ble.Scan(b.scanCtx, true, advHandler, filter)
	}()
}

func (b *bridge) startAPIService() {
	log.Infoln("Preparing api http services")

	router := mux.NewRouter()

	var validateToken = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.URL.Query()["token"]
			if !ok || len(token) != 1 || token[0] != b.token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				log.WithField("source", r.RemoteAddr).Warningln("Unauthorized request")
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	b.service = NewBridgeService(b)
	inofficialController := api.NewInofficialApiController(b.service)
	officialController := api.NewOfficialApiController(b.service)
	eventsController := api.NewEventsApiController(b.service)

	apiRouter := api.NewRouter(inofficialController, officialController, eventsController)
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	apiRouter.Use(validateToken)

	fileServer := http.FileServer(templates.Assets)

	router.PathPrefix("/auth").Handler(apiRouter)
	router.PathPrefix("/configAuth").Handler(apiRouter)
	router.PathPrefix("/list").Handler(apiRouter)
	router.PathPrefix("/lockState").Handler(apiRouter)
	router.PathPrefix("/lockAction").Handler(apiRouter)
	router.PathPrefix("/callback").Handler(apiRouter)
	router.PathPrefix("/locks").Handler(apiRouter)
	router.PathPrefix("/bridge").Handler(apiRouter)
	router.PathPrefix("/events").Handler(apiRouter)
	router.PathPrefix("/").Handler(fileServer)

	log.WithField("port", b.port).Infoln("serving web services")
	log.Fatal(http.ListenAndServe(b.port, router))
}
