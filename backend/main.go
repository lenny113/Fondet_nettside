package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	yfa "github.com/oscarli916/yahoo-finance-api"
)

type Holdings struct {
	Beholdning []map[string]float64 `json:"beholdning"`
}

type PortfolioEntry struct {
	FundID string  `json:"fund_id"`
	Units  float64 `json:"units"`
	Price  float64 `json:"price"`
	Value  float64 `json:"value"`
}

type PortfolioResponse struct {
	Holdings   []PortfolioEntry `json:"holdings"`
	TotalValue float64          `json:"total_value"`
}

func loadHoldings() (map[string]float64, error) {
	data, err := os.ReadFile("holdings.json")
	if err != nil {
		return nil, fmt.Errorf("reading holdings.json: %w", err)
	}
	var h Holdings
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("parsing holdings.json: %w", err)
	}
	agg := make(map[string]float64)
	for _, entry := range h.Beholdning {
		for id, amount := range entry {
			agg[id] += amount
		}
	}
	return agg, nil
}

func getLastValue(fundID string) (float64, error) {
	t := yfa.NewTicker(fundID)
	info, err := t.Info()
	if err != nil {
		return 0, fmt.Errorf("fetching %s: %w", fundID, err)
	}
	if info.RegularMarketPrice == nil {
		return 0, fmt.Errorf("no price for %s", fundID)
	}
	return info.RegularMarketPrice.Raw, nil
}

func handlePortfolio(w http.ResponseWriter, r *http.Request) {
	agg, err := loadHoldings()
	if err != nil {
		log.Printf("holdings error: %v", err)
		http.Error(w, "failed to load holdings", http.StatusInternalServerError)
		return
	}

	var resp PortfolioResponse
	for id, units := range agg {
		price, err := getLastValue(id)
		if err != nil {
			log.Printf("price error: %v", err)
			continue
		}
		value := units * price
		resp.TotalValue += value
		resp.Holdings = append(resp.Holdings, PortfolioEntry{
			FundID: id, Units: units, Price: price, Value: value,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/api/portfolio", handlePortfolio)
	log.Println("listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
