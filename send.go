package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ── Send ─────────────────────────────────────────────────────────────────────

func sendAllMonths(cfg Config) {
	for month := 1; month <= 12; month++ {
		file := fmt.Sprintf("%s/record_%d_%02d.json", cfg.OutputDir, year, month)

		logSection("month %02d  %s", month, file)

		if _, err := os.Stat(file); os.IsNotExist(err) {
			logSkip("file not found — skipping")
			fmt.Println()
			continue
		}

		if err := sendToSheets(cfg.SheetsURL, file, month); err != nil {
			logError("failed to send: %v", err)
		} else {
			logInfo("sent successfully")
		}

		fmt.Println()
	}
}

func sendToSheets(sheetsURL, file string, month int) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var filePayload SheetPayload
	if err := json.Unmarshal(data, &filePayload); err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	filePayload.Month = time.Month(month).String()

	payload := map[string]interface{}{
		"sheet": fmt.Sprintf("%d", year),
		"month": filePayload.Month,
		"data":  filePayload.Data,
	}

	finalBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	logInfo("payload: %s", string(finalBody))

	resp, err := http.Post(sheetsURL, "application/json", bytes.NewBuffer(finalBody))
	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logInfo("status : %s", resp.Status)
	logInfo("response: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}
