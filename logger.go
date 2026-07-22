package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// ── Globals ──────────────────────────────────────────────────────────────────

var logger *log.Logger
var debug string
var labelsFile string

var (
	walletToken  string
	sheetsURL    string
	outputDir    string
	outputLog    string
	configLabels string
	logLevel     string
)

// ── Logger ───────────────────────────────────────────────────────────────────

func initLogger(outputDir string) (*os.File, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log folder '%s': %w", outputDir, err)
	}
	logPath := fmt.Sprintf("%s/wallet_%s.log", outputDir, time.Now().Format("2006-01-02_15-04-05"))

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	multi := io.MultiWriter(os.Stdout, file)
	logger = log.New(multi, "", 0)

	logInfo("log file   : %s", logPath)

	return file, nil
}

func logSection(format string, args ...any) {
	fmt.Printf("──────────────────────────────────────\n")
	fmt.Printf("  "+format+"\n", args...)
	fmt.Printf("──────────────────────────────────────\n")
}

func logWarn(format string, args ...any) {
	logJSON("WARN", format, args...)
}

func logError(format string, args ...any) {
	logJSON("ERROR", format, args...)
}

func logSkip(format string, args ...any) {
	logJSON("SKIP", format, args...)
}

func logInfo(format string, args ...any) {
	logJSON("INFO", format, args...)
}

func logJSON(level string, format string, args ...any) {
	message := fmt.Sprintf(format, args...)

	pc, _, _, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()

	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"service":   os.Args[1],
		"function":  funcName,
		"message":   message,
	}

	b, _ := json.Marshal(entry)
	line := string(b)

	if logger != nil {
		logger.Print(line)
	} else {
		fmt.Println(line)
	}
}
