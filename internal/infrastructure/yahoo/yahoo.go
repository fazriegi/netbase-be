package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/shopspring/decimal"
)

const maxYahooResponseBytes = 2 << 20 // 2 MB

type yahooProvider struct {
	apiKey string
	client *http.Client
}

func NewYahooProvider(apiKey string) domain.YahooProvider {
	return &yahooProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type YahooFinanceResponse struct {
	OptionChain YahooFinanceOptionChain `json:"optionChain"`
}

type YahooFinanceOptionChain struct {
	Result []YahooFinanceResult `json:"result"`
}

type YahooFinanceResult struct {
	UnderlyingSymbol string            `json:"underlyingSymbol"`
	Quote            YahooFinanceQuote `json:"quote"`
}

type YahooFinanceQuote struct {
	RegularMarketPrice decimal.Decimal `json:"regularMarketPrice"`
}

func (y *yahooProvider) FetchPrice(ctx context.Context, ticker string) (decimal.Decimal, error) {
	escapedSymbol := url.QueryEscape(ticker + ".JK")
	endpoint := fmt.Sprintf("https://yahoo-finance-real-time1.p.rapidapi.com/stock/get-options?symbol=%s&lang=en-US", escapedSymbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return decimal.Zero, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-rapidapi-host", "yahoo-finance-real-time1.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", y.apiKey)

	res, err := y.client.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("stock price API returned status code %d", res.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, maxYahooResponseBytes))
	if err != nil {
		return decimal.Zero, err
	}

	var yahooFinanceResp YahooFinanceResponse
	err = json.Unmarshal(body, &yahooFinanceResp)
	if err != nil {
		return decimal.Zero, err
	}

	if len(yahooFinanceResp.OptionChain.Result) == 0 {
		return decimal.Zero, fmt.Errorf("no result found for ticker %s", ticker)
	}

	result := yahooFinanceResp.OptionChain.Result[0]
	if result.Quote.RegularMarketPrice.IsZero() {
		return decimal.Zero, fmt.Errorf("no market price found for ticker %s", ticker)
	}

	return result.Quote.RegularMarketPrice, nil
}
