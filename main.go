package main

import (
	"benchmark/logging"
	"benchmark/task"
	"log"
	"time"

	"github.com/EnclaveRunner/sdk-go/enclave"
	"github.com/google/uuid"
)

var (
	apiURL       = "http://localhost:8080"
	username     = "enclave"
	password     = "enclave"
	sampleSize   = 10000
	idleTime     = 0.5 // seconds
	receiverAddr = ":8083"
)

func main() {
	// create logger writing to logs.csv in the current directory
	logger, err := logging.NewLogger("logs.csv")
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	// start receiver HTTP server
	task.StartServer(receiverAddr, logger)

	// create enclave client
	client, err := enclave.NewClient(apiURL, username, password)
	if err != nil {
		panic(err)
	}

	// prepare suite (upload artifact if not exists)
	task.PrepareSuite(*client, logger)

	for i := 0; i < sampleSize; i++ {
		id := uuid.New().String()
		if err := task.CreateTask(id, *client, logger, receiverAddr); err != nil {
			logger.Error("CreateTask failed", logging.Fields{"task_id": id, "error": err.Error()})
			// continue publishing other tasks even if one fails
			continue
		}

		// idle between requests
		time.Sleep(time.Duration(idleTime) * time.Second)
	}

}
