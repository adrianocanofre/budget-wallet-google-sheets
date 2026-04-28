package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

func sumRecords(records []Record) float64 {
	var total float64
	for _, r := range records {
		total += r.Amount.Value
	}
	return total
}

func formatEuro(value float64) string {
	return fmt.Sprintf("%.2f", math.Abs(value))
}

func formatTotals(data map[string]float64) map[string]string {
	formatted := make(map[string]string, len(data))
	for label, value := range data {
		formatted[label] = formatEuro(value)
	}
	return formatted
}

func savePayload(payload SheetPayload, filename string) error {
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0644)
}

func monthRange(year, month int) (string, string) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}
