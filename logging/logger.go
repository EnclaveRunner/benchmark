package logging

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Fields are structured key/value pairs attached to a log entry.
type Fields map[string]interface{}

// Logger writes colored console logs via zerolog and structured rows to a CSV file.
type Logger struct {
	mu        sync.Mutex
	console   zerolog.Logger
	csvFile   *os.File
	csvWriter *csv.Writer
}

// NewLogger creates a Logger that writes to stdout (colored) and appends to the given CSV path.
// If the CSV file is empty a header row will be added.
func NewLogger(csvPath string) (*Logger, error) {
	f, err := os.OpenFile(csvPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	if fi, err := f.Stat(); err == nil && fi.Size() == 0 {
		if err := w.Write([]string{"timestamp", "level", "message", "fields"}); err != nil {
			f.Close()
			return nil, err
		}
		w.Flush()
	}

	// console writer with ANSI colors
	cw := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	cw.FormatLevel = func(i interface{}) string {
		s := fmt.Sprintf("%v", i)
		sl := strings.ToLower(s)
		color := ""
		reset := "\x1b[0m"
		switch {
		case strings.Contains(sl, "err"):
			color = "\x1b[31m"
		case strings.Contains(sl, "warn"):
			color = "\x1b[33m"
		case strings.Contains(sl, "info"):
			color = "\x1b[32m"
		case strings.Contains(sl, "debug"):
			color = "\x1b[34m"
		default:
			color = "\x1b[35m"
		}
		return fmt.Sprintf("%s%-6s%s", color, s, reset)
	}

	cz := zerolog.New(cw).With().Timestamp().Logger()

	return &Logger{
		console:   cz,
		csvFile:   f,
		csvWriter: w,
	}, nil
}

// Close flushes and closes the CSV file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.csvWriter != nil {
		l.csvWriter.Flush()
	}
	if l.csvFile != nil {
		return l.csvFile.Close()
	}
	return nil
}

func (l *Logger) writeCSV(timestamp, level, msg string, fields Fields) error {
	var jsonFields string
	if len(fields) > 0 {
		b, err := json.Marshal(fields)
		if err != nil {
			jsonFields = fmt.Sprintf(`{"_marshal_error":%q}`, err.Error())
		} else {
			jsonFields = string(b)
		}
	}
	if l.csvWriter == nil {
		return fmt.Errorf("csv writer not initialized")
	}
	if err := l.csvWriter.Write([]string{timestamp, level, msg, jsonFields}); err != nil {
		return err
	}
	l.csvWriter.Flush()
	return nil
}

// log writes a single entry to the colored console (zerolog) and CSV.
func (l *Logger) log(level, msg string, fields Fields) error {
	ts := time.Now().Format(time.RFC3339)
	l.mu.Lock()
	defer l.mu.Unlock()

	// console via zerolog events
	var ev *zerolog.Event
	switch strings.ToLower(level) {
	case "debug":
		ev = l.console.Debug()
	case "warn", "warning":
		ev = l.console.Warn()
	case "error", "err":
		ev = l.console.Error()
	default:
		ev = l.console.Info()
	}

	if len(fields) > 0 {
		// stable field order for console readability
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			ev.Interface(k, fields[k])
		}
	}
	ev.Msg(msg)

	if err := l.writeCSV(ts, strings.ToUpper(level), msg, fields); err != nil {
		// also print an inline warning to console
		l.console.Warn().Err(err).Msg("failed to write csv")
		return err
	}
	return nil
}

// Info logs an informational message with optional structured fields.
func (l *Logger) Info(msg string, fields Fields) error { return l.log("INFO", msg, fields) }

// Debug logs a debug message with optional structured fields.
func (l *Logger) Debug(msg string, fields Fields) error { return l.log("DEBUG", msg, fields) }

// Warn logs a warning message with optional structured fields.
func (l *Logger) Warn(msg string, fields Fields) error { return l.log("WARN", msg, fields) }

// Error logs an error message with optional structured fields.
func (l *Logger) Error(msg string, fields Fields) error { return l.log("ERROR", msg, fields) }
