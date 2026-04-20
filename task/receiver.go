package task

import (
	"benchmark/logging"
	"net/http"
	"time"
)

var (
	// map requestID to creation timestamp
	listeners = make(map[string]time.Time)
)

// CreateListener registers a listener timestamp for a request ID.
func CreateListener(requestID string, timestamp time.Time, logger *logging.Logger) {
	listeners[requestID] = timestamp
	logger.Info("listener registered", logging.Fields{"request": requestID, "timestamp": timestamp.Format(time.RFC3339)})
}

// StartServer starts an HTTP server that handles /benchmarks/?request=<id>
// It runs in a goroutine and returns immediately.
func StartServer(addr string, logger *logging.Logger) {
	http.HandleFunc("/benchmarks/", func(w http.ResponseWriter, r *http.Request) {
		reqID := r.URL.Query().Get("request")
		if reqID == "" {
			http.Error(w, "missing request param", http.StatusBadRequest)
			return
		}

		now := time.Now()
		logger.Info("benchmark endpoint called", logging.Fields{"request": reqID, "remote": r.RemoteAddr})

		if prev, ok := listeners[reqID]; ok {
			diff := now.Sub(prev)
			logger.Info("listener hit, reporting diff", logging.Fields{"request": reqID, "diff_ms": diff.Milliseconds()})
		} else {
			logger.Info("no previous listener found; registering new", logging.Fields{"request": reqID})
		}

		// register/update timestamp
		CreateListener(reqID, now, logger)

		w.WriteHeader(http.StatusNoContent)
	})

	go func() {
		logger.Info("starting benchmark HTTP server", logging.Fields{"addr": addr})
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("http server exited", logging.Fields{"error": err.Error()})
		}
	}()
}
