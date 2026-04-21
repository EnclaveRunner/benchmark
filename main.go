package main

import (
	"benchmark/logging"
	"benchmark/task"
	"log"

	"github.com/EnclaveRunner/sdk-go/enclave"
)

var (
	apiURL       = "http://localhost:8080"
	username     = "enclave"
	password     = "enclave"
	sampleSize   = 1000
	idleTime     = 0.5 // seconds
	receiverAddr = "localhost:8083"
)

func main() {
	// create logger writing to logs.csv in the current directory
	logger, err := logging.NewLogger("logs.csv")
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	// create enclave client
	client, err := enclave.NewClient(apiURL, username, password)
	if err != nil {
		panic(err)
	}

	// prepare suite (upload artifact if not exists)
	err = task.PrepareSuite(*client, logger)
	if err != nil {
		panic(err)
	}

	// start receiver HTTP server
	task.StartMeasuring(sampleSize, receiverAddr, logger, client)

	select {}
}
