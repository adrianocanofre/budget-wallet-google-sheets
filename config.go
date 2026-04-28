package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ── Constants ────────────────────────────────────────────────────────────────

const (
	baseURL = "https://rest.budgetbakers.com/wallet/v1/api/records"
	year    = 2026
)

// ── Config ───────────────────────────────────────────────────────────────────

func loadConfig() (Config, error) {
	token := os.Getenv("WALLET_TOKEN")
	if token == "" {
		token = walletToken
	}
	if token == "" {
		return Config{}, fmt.Errorf("WALLET_TOKEN env variable not set")
	}

	sheets := os.Getenv("SHEETS_URL")
	if sheets == "" {
		sheets = sheetsURL
	}
	if sheets == "" {
		return Config{}, fmt.Errorf("SHEETS_URL env variable not set")
	}

	outDir := os.Getenv("WALLET_OUTPUT_DIR")
	if outDir == "" && outputDir != "" {
		outDir = outputDir
	}
	if outDir == "" {
		outDir = "output"
	}

	outLog := os.Getenv("WALLET_OUTPUT_LOG")
	if outLog == "" && outputLog != "" {
		outLog = outputLog
	}
	if outLog == "" {
		outLog = "outputLog"
	}

	debug = os.Getenv("LOG_LEVEL")
	if debug == "" && logLevel != "" {
		debug = logLevel
	}
	if debug == "" {
		debug = "info"
	}

	labelsFile = os.Getenv("CONFIG_LABELS")
	if labelsFile == "" && configLabels != "" {
		labelsFile = configLabels
	}
	if labelsFile == "" {
		labelsFile = "config/labels.json"
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return Config{}, fmt.Errorf("failed to create output folder '%s': %w", outDir, err)
	}

	return Config{
		OutputDir: outDir,
		OutputLog: outLog,
		SheetsURL: sheets,
		Token:     token,
	}, nil
}

func loadLabels() (map[string]string, error) {
	data, err := os.ReadFile(labelsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", labelsFile, err)
	}

	var labels map[string]string
	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", labelsFile, err)
	}

	return labels, nil
}
