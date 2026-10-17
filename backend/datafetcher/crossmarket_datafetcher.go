package datafetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var DefaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Accept":     "application/json",
}

var APIDomainsList = []string{
	// --- Mainnet domains ---
	"https://api.bybit.com",  // Main domain - Singapore(AWS)
	"https://api.bytick.com", // Official fault tolerant mirror-domain

	// --- Regional mirrors ---
	"https://api.bybit-tr.com",      // Official domain for Turkey
	"https://api.bybit.nl",          // Official domain for Netherlands
	"https://api.bybit.kz",          // Official domain for Kazakhstan
	"https://api.bybitgeorgia.ge",   // Official domain for Georgia
	"https://api.spark-fintech.com", // Official domain for Hong Kong (Spark Fintech)
	"https://api.bybit.id",          // Dedicated domain for Indonesia
	"https://api.manepa.jp",         // Dedicated mirror-domain for Japan

	// --- Testnet & Demo ---
	"https://api-testnet.bybit.com", // Official public testnet endpoint
	"https://api-demo.bybit.com",    // General demotrading endpoint
}

type ConnectionSettings struct {
	BaseURL string
	Headers map[string]string
}

/*
NewConnectionSettings contains a required settings to proceed almost all tools
----------------------------------------------------------------------
That class is being used almost everywhere.
If it is not set then the default settings will be applied
-----------------------------------------------------------
base_url="https://bytick.com" # Official fault tolerant mirror-domain
DEFAULT_HEADERS = { # default headers that help to avoid cloudflare&AWS cdn blocks

	    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	    "Accept":     "application/json",
	}
*/
func NewConnectionSettings() ConnectionSettings {
	return ConnectionSettings{
		BaseURL: APIDomainsList[1], // Official fault tolerant mirror-domain
		Headers: DefaultHeaders,
	}
}

/*
CrossMarketDataFetcher is a struct to fetch market data (klines) from the Bybit API.

Attributes:

	ConnectionSettings (ConnectionSettings): Configuration settings containing the base URL.
	KlinesLimit (int): The maximum number of candlesticks (klines) to fetch per request.
*/
type CrossMarketDataFetcher struct {
	ConnectionSettings ConnectionSettings
	KlinesLimit        int
	HTTPClient         *http.Client
}

/*
NewCrossMarketDataFetcher initializes the CrossMarketDataFetcher with connection settings and request limits.

Args:

	settings (*ConnectionSettings): An instance of ConnectionSettings. Defaults to a new ConnectionSettings instance if nil.
	klinesLimit (int): The total amount of klines to retrieve. Defaults to 1000 if <= 0.
*/
func NewCrossMarketDataFetcher(settings *ConnectionSettings, klinesLimit int) *CrossMarketDataFetcher {
	s := NewConnectionSettings()
	if settings != nil {
		s = *settings
	}

	if klinesLimit <= 0 {
		klinesLimit = 1000
	}

	return &CrossMarketDataFetcher{
		ConnectionSettings: s,
		KlinesLimit:        klinesLimit,
		HTTPClient:         &http.Client{Timeout: 10 * time.Second},
	}
}

type BybitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List [][]string `json:"list"`
	} `json:"result"`
}

/*
GetKlines fetches historical klines (candlestick data) for a specific symbol from Bybit.

Args:

	symbol (string): The trading pair symbol. Defaults to "BTCUSDT" if empty.
	period (string): The candlestick chart interval (e.g., "D" for daily, "1" for 1m). Defaults to "D" if empty.
	marketType (string): The market category (e.g., "spot", "linear", "inverse"). Defaults to "spot" if empty.

Raises:

	ValueError: If the Bybit API returns a non-zero response code (indicating an error) or HTTP status is invalid.

Returns:

	[][]string: A list of klines data if the request is successful; otherwise, an empty list.
*/
func (f *CrossMarketDataFetcher) GetKlines(symbol, period, marketType string) [][]string {
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if period == "" {
		period = "D"
	}
	if marketType == "" {
		marketType = "spot"
	}

	endpoint := "/v5/market/kline"
	baseURL := strings.TrimRight(f.ConnectionSettings.BaseURL, "/")
	fullURL := baseURL + endpoint

	u, err := url.Parse(fullURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return [][]string{}
	}

	q := u.Query()
	q.Set("category", marketType)
	q.Set("symbol", symbol)
	q.Set("interval", period)
	q.Set("limit", strconv.Itoa(f.KlinesLimit)) // defines the total amount of klines
	u.RawQuery = q.Encode()

	// Idiomatic Go error handling instead of try-except block
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		fmt.Printf("Error fetching data for %s: %v\n", symbol, err)
		return [][]string{}
	}

	for key, value := range f.ConnectionSettings.Headers {
		req.Header.Set(key, value)
	}

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching data for %s: %v\n", symbol, err)
		return [][]string{}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error fetching data for %s: %v\n", symbol, err)
		return [][]string{}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview := string(bodyBytes)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		fmt.Printf("Error fetching data for %s: HTTP Status %d. Preview: %s\n", symbol, resp.StatusCode, preview)
		return [][]string{}
	}

	var apiResp BybitResponse
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&apiResp); err != nil {
		preview := string(bodyBytes)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		fmt.Printf("Error fetching data for %s: %v. Raw response was: %s\n", symbol, err, preview)
		return [][]string{}
	}

	if apiResp.RetCode != 0 {
		fmt.Printf("Error fetching data for %s: Bybit API Error: %s\n", symbol, apiResp.RetMsg)
		return [][]string{}
	}

	return apiResp.Result.List
}
