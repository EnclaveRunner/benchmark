package task

import (
	"benchmark/logging"
	"net/http"
	"time"
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
func StartServer(addr string, logger *logging.Logger) {
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
				"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
			})
			w.WriteHeader(http.StatusNoContent)
		})
	}

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("http server exited", logging.Fields{"error": err.Error()})
		}
	}()
}
