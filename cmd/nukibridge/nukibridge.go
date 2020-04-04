package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"

	"github.com/mapero/nuki-bridge/pkg/nukibridge"
	log "github.com/sirupsen/logrus"
)

var (
	verboseFlag    = flag.Bool("verbose", false, "verbose log entries for debugging")
	tokenFlag      = flag.String("token", "", "authentication token for api calls")
	configPathFlag = flag.String("config", "", "configuration path")
	portFlag       = flag.String("port", ":8080", "api port")
	done           = make(chan struct{})
)

func main() {
	flag.Parse()

	if *verboseFlag {
		log.SetLevel(log.DebugLevel)
	}

	token, ok := os.LookupEnv("NUKI_TOKEN")
	if !ok {
		if *tokenFlag != "" {
			token = *tokenFlag
		} else {
			var randToken [16]byte
			_, err := rand.Read(randToken[:])
			if err != nil {
				log.WithError(err).Fatalln("Failed to generate token")
			}
			token = fmt.Sprintf("%x", randToken)
			log.WithField("token", token).Infoln("Generated token for api")
		}
	}

	configPath, ok := os.LookupEnv("NUKI_CONFIGPATH")
	if !ok {
		if *configPathFlag != "" {
			configPath = *configPathFlag
		} else {
			configPath, _ = os.Getwd()
		}
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = *portFlag
	}

	_, err := nukibridge.NewBridge(configPath, port, token)
	if err != nil {
		panic(err)
	}

	<-done
	log.Infoln("Done")
}
