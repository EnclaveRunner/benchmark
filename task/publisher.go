package task

import (
	"benchmark/logging"
	"context"
	"os"
	"time"

	"github.com/EnclaveRunner/sdk-go/enclave"
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

func PrepareSuite(client enclave.Client, logger *logging.Logger) {
	logger.Info("Preparing benchmark suite", logging.Fields{
		"namespace": benchmarkNamespace,
		"name":      benchmarkName,
	})

	_, err := client.GetArtifactByTag(context.Background(), benchmarkNamespace, benchmarkName, benchmarkTag)
	if err == nil {
		logger.Info("Benchmark artifact already exists", logging.Fields{
			"namespace": benchmarkNamespace,
			"name":      benchmarkName,
		})
		return
	}

	logger.Info("Benchmark artifact not found, uploading from local file", logging.Fields{
		"path":  "./benchmark.wasm",
		"error": err.Error(),
	})

	f, err := os.Open("./benchmark.wasm")
	if err != nil {
		logger.Error("Failed to open benchmark artifact file", logging.Fields{
			"path":  "./benchmark.wasm",
			"error": err.Error(),
		})
		return
	}
	defer f.Close()

	uploadedArtifact, err := client.UploadArtifactRaw(context.Background(), benchmarkNamespace, benchmarkName, benchmarkTag, f)
	if err != nil {
		logger.Error("Failed to upload benchmark artifact", logging.Fields{
			"namespace": benchmarkNamespace,
			"name":      benchmarkName,
			"error":     err.Error(),
		})
		return
	}

	logger.Info("Benchmark artifact uploaded", logging.Fields{
		"namespace": benchmarkNamespace,
		"name":      benchmarkName,
	})

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
	}
}

func CreateTask(id string, client enclave.Client, logger *logging.Logger, receiverAddr string) error {
	req := defaultTask
	callback := "/benchmarks/?request=" + id
	req.Params = []any{"http://localhost" + receiverAddr + callback}
	_, err := client.CreateTask(context.Background(), req)
	timestamp := time.Now()
	logger.Info("Task published", logging.Fields{"task_id": id, "timestamp": timestamp, "callback": callback})
	CreateListener(id, timestamp, logger)
	return err
}
