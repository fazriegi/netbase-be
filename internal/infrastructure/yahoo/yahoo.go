package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fazriegi/netbase-be/internal/domain"
	"github.com/shopspring/decimal"
)

type yahooProvider struct {
	apiKey string
}

func NewYahooProvider(apiKey string) domain.YahooProvider {
	return &yahooProvider{apiKey: apiKey}
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
	url := fmt.Sprintf("https://yahoo-finance-real-time1.p.rapidapi.com/stock/get-options?symbol=%s.JK&lang=en-US", ticker)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return decimal.Zero, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-rapidapi-host", "yahoo-finance-real-time1.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", y.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("stock price API returned status code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
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
