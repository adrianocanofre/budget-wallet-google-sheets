package main

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
