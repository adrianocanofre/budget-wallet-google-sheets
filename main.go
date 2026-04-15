package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"time"
)

// ── Types ────────────────────────────────────────────────────────────────────

type Amount struct {
	Value float64 `json:"value"`
}

type Label struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Record struct {
	Amount Amount  `json:"amount"`
	Labels []Label `json:"labels"`
	Note   string  `json:"note"`
}

type APIResponse struct {
	Records []Record `json:"records"`
}

type SheetPayload struct {
	Month string            `json:"month"`
	Data  map[string]string `json:"data"`
}

type Config struct {
	OutputDir string
	SheetsURL string
	Token     string
	OutputLog string
}

// ── Constants ────────────────────────────────────────────────────────────────

const (
	baseURL    = "https://rest.budgetbakers.com/wallet/v1/api/records"
	year       = 2026
	labelsFile = "config/labels.json"
)

var logger *log.Logger
var debug string

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

// ── Bootstrap ────────────────────────────────────────────────────────────────

func loadConfig() (Config, error) {
	token := os.Getenv("WALLET_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("WALLET_TOKEN env variable not set")
	}

	sheetsURL := os.Getenv("SHEETS_URL")
	if sheetsURL == "" {
		return Config{}, fmt.Errorf("SHEETS_URL env variable not set")
	}

	outputDir := os.Getenv("WALLET_OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "output"
	}

	outputLog := os.Getenv("WALLET_OUTPUT_LOG")
	if outputLog == "" {
		outputLog = "outputLog"
	}

	debug = os.Getenv("LOG_LEVEL")
	if debug == "" {
		debug = "info"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return Config{}, fmt.Errorf("failed to create output folder '%s': %w", outputDir, err)
	}

	return Config{
		OutputDir: outputDir,
		OutputLog: outputLog,
		SheetsURL: sheetsURL,
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("go run main.go fetch   → fetch all months and save JSON files")
		fmt.Println("go run main.go send    → send all saved JSON files to Google Sheets")
		return
	}
	fmt.Println(debug)
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("[ERROR] config: %v", err)
	}
	logFile, err := initLogger(cfg.OutputLog)
	if err != nil {
		log.Fatalf("[ERROR] logger: %v", err)
	}
	defer logFile.Close()

	logInfo("output dir : %s", cfg.OutputDir)
	logInfo("sheets url : %s", cfg.SheetsURL)

	switch os.Args[1] {
	case "fetch":
		fetchAllMonths(cfg)
	case "send":
		sendAllMonths(cfg)
	default:
		log.Fatalf("[ERROR] unknown command '%s' — use: fetch | send", os.Args[1])
	}
}

// ── Fetch ────────────────────────────────────────────────────────────────────

func fetchAllMonths(cfg Config) {
	labels, err := loadLabels()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	logInfo("loaded %d labels from %s", len(labels), labelsFile)

	for month := 1; month <= 12; month++ {

		dateFrom, dateTo := monthRange(year, month)
		outputFile := fmt.Sprintf("%s/record_%d_%02d.json", cfg.OutputDir, year, month)

		logSection("month %02d  %s → %s", month, dateFrom, dateTo)

		totals := make(map[string]float64)
		totalOk := 0
		totalSkip := 0

		for name, id := range labels {
			records, err := fetchRecordsByLabel(cfg.Token, dateFrom, dateTo, id)
			if err != nil {
				logWarn("label %s error: %v", "'"+name+"'", err)
				continue
			}
			sum := sumRecords(records)
			if sum != 0 {
				totals[name] = sum
				totalOk++
				logInfo("%s=%s", name, formatEuro(sum))
			} else {
				totals[name] = 0
				totalSkip++
				logSkip("%s no records", name)
			}
			time.Sleep(300 * time.Millisecond)
		}

		if len(totals) == 0 {
			logWarn("no records for month %02d — skipping", month)
			fmt.Println()
			continue
		}

		payload := SheetPayload{
			Month: time.Month(month).String(),
			Data:  formatTotals(totals),
		}

		if err := savePayload(payload, outputFile); err != nil {
			logError("failed to save file: %v", err)
			continue
		}

		logInfo("total Label → %d", totalOk)
		logInfo("total no records → %d", totalSkip)
		logInfo("saved → %s", outputFile)
		fmt.Println()
	}
}

func fetchRecordsByLabel(token, dateFrom, dateTo, labelID string) ([]Record, error) {
	const pageSize = 100
	var allRecords []Record
	offset := 0

	for {
		url := fmt.Sprintf(
			"%s?recordDate=gte.%s%%2Clt.%s&labelId=%s&limit=%d&offset=%d&agentHints=false",
			baseURL, dateFrom, dateTo, labelID, pageSize, offset,
		)

		logInfo("GET offset=%d url=%s", offset, url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		allRecords = append(allRecords, apiResp.Records...)

		logInfo("fetched %d records at offset %d", len(apiResp.Records), offset)

		// if we got fewer than pageSize, we've reached the last page
		if len(apiResp.Records) < pageSize {
			break
		}

		offset += pageSize
	}

	return allRecords, nil
}

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
	if debug == "debug" {
		logJSON("INFO", format, args...)
	} else {
		return
	}

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
