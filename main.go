package main

import (
	"github.com/EnclaveRunner/sdk-go/enclave"
	"github.com/google/uuid"
)

var (
	apiURL     = "http://localhost:8080"
	username   = "enclave"
	password   = "enclave"
	sampleSize = 100
)

func main() {
	client, err := enclave.NewClient(apiURL, username, password)
	if err != nil {
		panic(err)
	}

	for i := 0; i < sampleSize; i++ {
		id := uuid.New().String()
		if err := createTask(id, client); err != nil {
			panic(err)
		}
	}

}
