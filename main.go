package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("go run . fetch   → fetch all months and save JSON files")
		fmt.Println("go run . send    → send all saved JSON files to Google Sheets")
		return
	}

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
