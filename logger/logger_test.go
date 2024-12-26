package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Test default configuration
	logger := NewLogger(Config{})
	if logger.level != DEBUG {
		t.Errorf("Expected default level to be DEBUG, got %v", logger.level)
	}
	if len(logger.outputs) != 1 {
		t.Errorf("Expected 1 default output, got %d", len(logger.outputs))
	}
	if logger.callDepth != 2 {
		t.Errorf("Expected default call depth to be 2, got %d", logger.callDepth)
	}
}

func TestTextFormatter(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Outputs: []LogOutput{{
			Writer:    &buf,
			Formatter: &TextFormatter{TimeFormat: time.RFC3339},
		}},
	})

	testMessage := "test message"
	testFields := map[string]interface{}{"key": "value"}
	logger.Info(testMessage, testFields)

	output := buf.String()
	if !strings.Contains(output, testMessage) {
		t.Errorf("Expected output to contain message '%s', got '%s'", testMessage, output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected output to contain field 'key=value', got '%s'", output)
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Outputs: []LogOutput{{
			Writer:    &buf,
			Formatter: &JSONFormatter{TimeFormat: time.RFC3339},
		}},
	})

	testMessage := "test message"
	testFields := map[string]interface{}{"key": "value"}
	logger.Info(testMessage, testFields)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if logEntry["message"] != testMessage {
		t.Errorf("Expected message '%s', got '%v'", testMessage, logEntry["message"])
	}
	if fields, ok := logEntry["fields"].(map[string]interface{}); !ok || fields["key"] != "value" {
		t.Errorf("Expected fields to contain {'key': 'value'}, got %v", logEntry["fields"])
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level: INFO,
		Outputs: []LogOutput{{
			Writer:    &buf,
			Formatter: &TextFormatter{TimeFormat: time.RFC3339},
		}},
	})

	tests := []struct {
		level   LogLevel
		message string
		should  bool
	}{
		{DEBUG, "debug message", false},
		{INFO, "info message", true},
		{WARN, "warn message", true},
		{ERROR, "error message", true},
	}

	for _, tt := range tests {
		buf.Reset()
		switch tt.level {
		case DEBUG:
			logger.Debug(tt.message, nil)
		case INFO:
			logger.Info(tt.message, nil)
		case WARN:
			logger.Warn(tt.message, nil)
		case ERROR:
			logger.Error(tt.message, nil)
		}

		if tt.should && !strings.Contains(buf.String(), tt.message) {
			t.Errorf("Expected output to contain '%s' for level %v", tt.message, tt.level)
		}
		if !tt.should && strings.Contains(buf.String(), tt.message) {
			t.Errorf("Expected output not to contain '%s' for level %v", tt.message, tt.level)
		}
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Outputs: []LogOutput{{
			Writer:    &buf,
			Formatter: &TextFormatter{TimeFormat: time.RFC3339},
		}},
	})

	context := map[string]interface{}{"context_key": "context_value"}
	contextLogger := logger.WithContext(context)

	testMessage := "test message"
	testFields := map[string]interface{}{"field_key": "field_value"}
	contextLogger.Info(testMessage, testFields)

	output := buf.String()
	if !strings.Contains(output, "context_key=context_value") {
		t.Errorf("Expected output to contain context 'context_key=context_value', got '%s'", output)
	}
	if !strings.Contains(output, "field_key=field_value") {
		t.Errorf("Expected output to contain field 'field_key=field_value', got '%s'", output)
	}
}
