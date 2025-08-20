package output_test

import (
	"testing"

	"github.com/mona-actions/gh-bbs-analyzer/internal/output"
)

func TestNewLogger(t *testing.T) {
	logger := output.NewLogger()
	if logger == nil {
		t.Error("Expected logger to be created")
	}

	err := logger.Close()
	if err != nil {
		t.Errorf("Expected no error closing logger, got %v", err)
	}
}

func TestLogger_Info(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test that Info method doesn't panic
	testMessage := "This is a test info message"
	logger.Info(testMessage)
	// Since we're only outputting to console now, we can't easily test the output
	// But we can ensure the method executes without error
}

func TestLogger_Warning(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test that Warning method doesn't panic
	testMessage := "This is a test warning"
	logger.Warning(testMessage)
}

func TestLogger_Error(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test that Error method doesn't panic
	testMessage := "This is a test error"
	logger.Error(testMessage)
}

func TestLogger_LogConfiguration(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test that LogConfiguration method doesn't panic
	key := "Test Key"
	value := "Test Value"
	logger.LogConfiguration(key, value)
}

func TestLogger_LogStatistic(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test that LogStatistic method doesn't panic
	key := "Test Stat"
	value := "Test Value"
	logger.LogStatistic(key, value)
}

func TestLogger_Debug(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test that Debug method doesn't panic
	testMessage := "This is a test debug message"
	logger.Debug(testMessage)
}

func TestLogger_CloseMultipleTimes(t *testing.T) {
	logger := output.NewLogger()

	// Close multiple times should not cause error
	err1 := logger.Close()
	err2 := logger.Close()

	if err1 != nil {
		t.Errorf("Expected no error on first close, got %v", err1)
	}

	if err2 != nil {
		t.Errorf("Expected no error on second close, got %v", err2)
	}
}

func TestLogger_EmptyMessage(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test with empty message - should not panic
	logger.Info("")
	logger.Debug("")
	logger.Warning("")
	logger.Error("")
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	logger := output.NewLogger()
	defer logger.Close()

	// Test concurrent writes - should not panic
	done := make(chan bool, 3)

	go func() {
		logger.Info("Message from goroutine 1")
		done <- true
	}()

	go func() {
		logger.Warning("Message from goroutine 2")
		done <- true
	}()

	go func() {
		logger.Error("Message from goroutine 3")
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}
