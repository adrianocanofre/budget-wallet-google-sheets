package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func fetchAllMonths(cfg Config, month_start int, printOnly bool) {

	monthFinish := int(time.Now().Month())
	if month_start >= 1 {
		monthFinish = month_start
	}
	labels, err := loadLabels()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	logInfo("loaded %d labels from %s", len(labels), labelsFile)

	for month := month_start; month <= monthFinish; month++ {

		dateFrom, dateTo := monthRange(year, month)
		outputFile := fmt.Sprintf("%s/record_%d_%02d.json", cfg.OutputDir, year, month)

		logSection("month %02d  %s → %s", month, dateFrom, dateTo)
		logInfo("month:  %02d", month)
		totals := make(map[string]float64)
		totalOk := 0
		totalSkip := 0

		for name, id := range labels {
			logSection("Label:  %s,  ID: %s", name, id)
			logInfo("label: %s", name)
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
				logInfo("%s=%s", name, formatEuro(0.00))
				logSkip("%s no records", name)
			}
			time.Sleep(300 * time.Millisecond)

			logSection("Label:  %s, value: %f", name, sum)
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

		if printOnly {
			b, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				logError("failed to marshal payload: %v", err)
				continue
			}

			fmt.Println(string(b))
		} else {
			if err := savePayload(payload, outputFile); err != nil {
				logError("failed to save file: %v", err)
				continue
			}

			logInfo("saved → %s", outputFile)
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

		logInfo("Status: %d", res.StatusCode)
		logInfo("Content-Type: %s", res.Header.Get("Content-Type"))
		logInfo("Response Body:\n%s", string(body))

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"API returned %d\nBody: %s",
				res.StatusCode,
				string(body),
			)
		}

		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf(
				"failed to parse response: %w\nBody: %s",
				err,
				string(body),
			)
		}

		allRecords = append(allRecords, apiResp.Records...)

		logInfo("fetched %d records at offset %d", len(apiResp.Records), offset)

		if len(apiResp.Records) < pageSize {
			break
		}

		offset += pageSize
	}

	return allRecords, nil
}
