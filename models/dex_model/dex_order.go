package dex_model

import (
	"github.com/shopspring/decimal"
	"time"
)

const (
	Buy  = "buy"
	Sell = "sell"

	Limit  = "limit"
	Market = "market"
	Cancel = "cancel"

	TimeInForceGTC = "GTC" // 订单一直有效，知道被成交或者取消
	TimeInForceIOC = "IOC" // 无法立即成交的部分就撤销
	TimeInForceFOK = "FOK" // 无法全部立即成交就撤销

	UnKnown    = 0 // 未知
	NotTraded  = 1 // 未成交
	PartTraded = 2 // 部分成交
	AllTraded  = 3 // 全部成交
	Canceled   = 4 //已撤单
	Rejected   = 5 // 已拒绝
)

func GetSideType(side string) int {
	if side == Buy {
		return 0
	} else {
		return 1
	}
}

func GetOrderType(orderType string) int {
	if orderType == Limit {
		return 0
	} else if orderType == Market {
		return 1
	} else if orderType == Cancel {
		return 2
	} else {
		return 3
	}
}

type DexOrder struct {
	ID        uint64          `json:"id" gorm:"primaryKey"`
	UId       uint64          `json:"user_id" gorm:"column:uid"`
	OrderID   string          `json:"order_id" gorm:"column:order_no"`
	Symbol    string          `json:"symbol" gorm:"column:symbol"`
	SideType  int             `json:"side_type" gorm:"column:side_type"`
	OrderType int             `json:"order_type" gorm:"column:order_type"`
	Price     decimal.Decimal `json:"price" gorm:"column:price"`
	// 总成交数量
	TotalVolume decimal.Decimal `json:"total_volume" gorm:"column:total_volume"`
	// 已经成交数量
	TradedVolume decimal.Decimal `json:"traded_volume" gorm:"column:traded_volume"`
	// 已成交金额 = 成交均价 * 成交数量
	TradedAmount decimal.Decimal `json:"traded_amount" gorm:"column:traded_amount"`
	Status       int             `json:"status" gorm:"column:status"`
	FeeAmount    decimal.Decimal `json:"fee_amount" gorm:"column:fee_amount"`
	FeeCoinName  string          `json:"fee_coin_name" gorm:"column:fee_coin_name"`
	CreateTime   time.Time       `json:"create_time" gorm:"column:create_time"`
	UpdateTime   time.Time       `json:"update_time" gorm:"column:update_time"`
}
