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
	_, err := client.GetArtifactByTag(context.Background(), benchmarkNamespace, benchmarkName, benchmarkTag)
	if err == nil {
		return
	}

	f, err := os.Open("./benchmark.wasm")
	if err != nil {
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
	task, err := client.CreateTask(context.Background(), req)
	timestamp := time.Now()
	if err != nil {
		return err
	}
	logger.Info("Task published", logging.Fields{"request": id, "callback": callback})
	CreateListener(id, timestamp, logger)
	completed := false
	for !completed && time.Since(timestamp) < 30*time.Second {
		currentTask, pollErr := client.GetTask(context.Background(), task.ID)
		if pollErr != nil {
			logger.Error("Failed to get task", logging.Fields{"task_id": task.ID, "error": pollErr.Error()})
			return pollErr
		}
		if currentTask.Status.State == "completed" {
			completed = true
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
