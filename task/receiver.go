package task

import (
	"benchmark/logging"
	"net/http"
	"os"
	"time"

	"github.com/EnclaveRunner/sdk-go/enclave"
)

// StartServer starts an HTTP server with the following endpoints:
//
//	/benchmarks/assigned/?request=<id>
//	/benchmarks/pulled/?request=<id>
//	/benchmarks/started/?request=<id>
//	/benchmarks/cleanup/?request=<id>
//
// Each endpoint logs the request ID and current timestamp, then returns 204.
// It runs in a goroutine and returns immediately.
func StartMeasuring(sampleSize int, addr string, logger *logging.Logger, client *enclave.Client) {
	samples := 0
	refTime := time.Now()

	for _, event := range []string{"assigned", "pulled", "started", "cleanup"} {
		http.HandleFunc("/benchmarks/"+event+"/", func(w http.ResponseWriter, r *http.Request) {
			reqID := r.URL.Query().Get("request")
			if reqID == "" {
				http.Error(w, "missing request param", http.StatusBadRequest)
				return
			}
			logger.Info("benchmark event", logging.Fields{
				"event":     event,
				"id":        reqID,
				"timestamp": time.Since(refTime).Nanoseconds(),
			})
			w.WriteHeader(http.StatusNoContent)
			if event == "cleanup" {
				samples++
				if samples >= sampleSize {
					logger.Info("benchmark completed, exiting", logging.Fields{})
					os.Exit(0)
				}

				err := CreateTask(client, logger, addr, refTime)
				if err != nil {
					logger.Info("failed to create task", logging.Fields{"error": err.Error()})
					os.Exit(1)
				}
				
				logger.Info("starting next benchmark", logging.Fields{"samples_completed": samples})
			}
		})
	}

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("http server exited", logging.Fields{"error": err.Error()})
		}
	}()

	time.Sleep(1 * time.Second) // Wait for http server to start

	err := CreateTask(client, logger, addr, refTime)
	if err != nil {
		logger.Info("failed to create task", logging.Fields{"error": err.Error()})
		os.Exit(1)
	}
}
