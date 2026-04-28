# budget-wallet-google-sheets
APP to integration budgetBakers to Google Sheets


## Build

```
GOOS=linux GOARCH=amd64 go build \
  -ldflags " \
    -X main.walletToken=<TOKEN>\
    -X main.sheetsURL=<GOOGLE_SHEETS_URL>\
    -X main.logLevel=debug \
    -X main.configLabels=/etc/wallet/config.json \
    -X main.outputDir=/etc/wallet/2026 \
    -X main.outputLog=/var/log/wallet" \
  -o wallet *.go
```

## Googhe Sheets

