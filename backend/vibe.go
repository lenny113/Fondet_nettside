package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Portfolio struct {
	Name   string
	Ticker string
	Shares float64
	Value  float64
}

func fetchPrice(symbol string) (float64, error) {

	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s.us&f=sd2t2ohlcv&h&e=csv", strings.ToLower(symbol))

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	data, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	if len(data) < 2 {
		return 0, fmt.Errorf("ingen data for %s", symbol)
	}

	priceStr := data[1][6]

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, err
	}

	return price, nil
}

func fetchUSDtoNOK() (float64, error) {

	url := "https://stooq.com/q/l/?s=usdnok&f=sd2t2ohlcv&h&e=csv"

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)

	data, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	if len(data) < 2 {
		return 0, fmt.Errorf("ingen data for USD/NOK")
	}

	priceStr := data[1][6]

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, err
	}

	return price, nil
}

func main() {

	funds := []Portfolio{
		{"Nordnet Teknologi Indeks NOK", "QQQ", 177.6604, 0},
		{"Nordnet Global Indeks NOK", "VT", 1280.1443, 0},
		{"Nordnet Europa Indeks NOK", "VGK", 133.7084, 0},
		{"Nordnet Emerging Markets Indeks", "VWO", 216.8626, 0},
		{"Kraft Corporate Bonds B", "BND", 489.536, 0},
		{"DNB Finans A", "XLF", 8.7997, 0},
	}

	usdToNok, err := fetchUSDtoNOK()
	if err != nil {
		panic(err)
	}

	fmt.Printf("USD/NOK kurs: %.2f\n\n", usdToNok)

	var totalValue float64

	fmt.Printf("=== PORTFØLJE %s ===\n\n", time.Now().Format("02.01 15:04"))

	for i := range funds {

		priceUSD, err := fetchPrice(funds[i].Ticker)

		if err != nil {
			fmt.Printf("❌ %s (%s): %v\n", funds[i].Name, funds[i].Ticker, err)
			continue
		}

		priceNOK := priceUSD * usdToNok

		value := priceNOK * funds[i].Shares
		funds[i].Value = value
		totalValue += value

		fmt.Printf("✅ %s\n", funds[i].Name)
		fmt.Printf("   Pris: %.2f USD (%.2f NOK)\n", priceUSD, priceNOK)
		fmt.Printf("   %.4f andeler = %.0f NOK\n\n",
			funds[i].Shares,
			value,
		)
	}

	fmt.Printf("💰 TOTAL PORTFØLJEVERDI: %.0f NOK\n", totalValue)
	fmt.Println("=====================================================")
}
