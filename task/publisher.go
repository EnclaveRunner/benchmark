package task

import (
	"benchmark/logging"
	"context"
	"errors"
	"os"
	"time"

	"github.com/EnclaveRunner/sdk-go/enclave"
	"github.com/google/uuid"
)

var (
	benchmarkName      = "benchmark"
	benchmarkNamespace = "enclave"
	functionName       = "startuptime"
	interfaceName      = "test-suite"
	benchmarkTag       = "benchmark"
)

var defaultTask = enclave.CreateTaskRequest{
	Source:   benchmarkNamespace + ":" + benchmarkName + "/" + interfaceName + "/" + functionName + "@" + benchmarkTag,
	Callback: "",
	Retries:  0,
}

var defaultTaskDocker = enclave.CreateTaskRequest{
	Source:   benchmarkNamespace + ":" + benchmarkName + "/" + "interface" + "/" + "ghcr.io/enclaverunner/benchmark:latest" + "@" + benchmarkTag,
	Callback: "",
	Retries:  0,
}

func PrepareSuite(client enclave.Client, logger *logging.Logger) error {
	_, err := client.GetArtifactByTag(context.Background(), benchmarkNamespace, benchmarkName, benchmarkTag)
	if err == nil {
		return nil
	}

	if !errors.Is(err, enclave.ErrNotFound) {
		logger.Error("Failed to check existing benchmark artifact", logging.Fields{
			"namespace": benchmarkNamespace,
			"name":      benchmarkName,
			"error":     err.Error(),
		})
		return err
	}

	f, err := os.Open("artifact/benchmark.wasm")
	if err != nil {
		return err
	}
	defer f.Close()

	uploadedArtifact, err := client.UploadArtifactRaw(context.Background(), benchmarkNamespace, benchmarkName, benchmarkTag, f)
	if err != nil {
		logger.Error("Failed to upload benchmark artifact", logging.Fields{
			"namespace": benchmarkNamespace,
			"name":      benchmarkName,
			"error":     err.Error(),
		})
		return err
	}

	patchReq := enclave.PatchArtifactRequest{
		Tags: []string{"benchmark"},
	}

	_, err = client.PatchArtifactByHash(context.Background(), benchmarkNamespace, benchmarkName, uploadedArtifact.VersionHash, patchReq)
	if err != nil {
		logger.Error("Failed to patch uploaded artifact", logging.Fields{
			"namespace": benchmarkNamespace,
			"name":      benchmarkName,
			"error":     err.Error(),
		})
		return err
	}

	return nil
}

func CreateTask(client *enclave.Client, logger *logging.Logger, receiverAddr string, refTime time.Time) error {
	id := uuid.NewString()
	req := defaultTaskDocker
	req.Env = []enclave.EnvironmentVariable{
		{
			Key:   "MEASUREMENT_ID",
			Value: id,
		},
		{
			Key:   "MEASUREMENT_SERVER",
			Value: "http://" + receiverAddr,
		},
	}

	_, err := client.CreateTask(context.Background(), req)
	if err != nil {
		return err
	}
	logger.Info("benchmark event", logging.Fields{
		"event":     "published",
		"id":        id,
		"timestamp": time.Now().Sub(refTime).Nanoseconds(),
	})
	return nil
}
