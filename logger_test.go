package http_logger_test

import (
	"github.com/azvaliev/http-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

type testProperty struct {
	Key      string
	Expected string
}

func TestWithLogging(t *testing.T) {
	logger, logs := setupLogsCapture()

	// Test case data
	testMethod := testProperty{"method", "POST"}
	testPath := testProperty{"path", "/api/data"}

	// Built into httptest.NewRequest
	testProto := testProperty{"proto", "HTTP/1.1"}
	testRemoteAddr := testProperty{"remoteAddr", "192.0.2.1:1234"}

	testProperties := []testProperty{testMethod, testPath, testProto, testRemoteAddr}

	// Create a simple test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Success!"))
	})

	// Wrap the test handler with our logging middleware
	handlerWithLogging := http_logger.WithLogging(testHandler, logger)

	// Create a test request
	req := httptest.NewRequest(testMethod.Expected, testPath.Expected, nil)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	handlerWithLogging.ServeHTTP(rr, req)

	// Get the log, make sure nothing additional was logged
	if len(logs.All()) != 1 {
		t.Error("Expected 1 log entry, got", len(logs.All()))
	}
	logEntry := logs.All()[0]
	if logEntry.Level != zap.InfoLevel {
		t.Error("Expected log level Info, got", logEntry.Level)
	}

	// Check the log details
	logDetails := logEntry.ContextMap()
	if logDetails["status"] != int64(http.StatusCreated) {
		t.Errorf("Expected status %d, got %s", http.StatusCreated, logDetails["status"])
	}
	for _, prop := range testProperties {
		if logDetails[prop.Key] != prop.Expected {
			t.Errorf("Expected %s %s, got %s", prop.Key, prop.Expected, logDetails[prop.Key])
		}
	}
}

func ExampleWithLogging() {
	// Create a zap test logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Success!"))
	})

	// Wrap the test handler with our logging middleware
	http.ListenAndServe(":8080", http_logger.WithLogging(http.DefaultServeMux, logger))
}
