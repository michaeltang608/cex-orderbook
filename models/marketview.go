package models

import (
	"github.com/shopspring/decimal"
)

// MarketView represents order book in a glance
type MarketView struct {
	Symbol string                     `json:"symbol"`
	Asks   map[string]decimal.Decimal `json:"asks"`
	Bids   map[string]decimal.Decimal `json:"bids"`
}
