package store

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order structure
type Order struct {
	CalendarID  string    `json:"calendarID"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	OperID      string    `json:"operID"`
	EventID     string    `json:"eventID"`
	ColorId     string    `json:"colorid"`
	FileURL     string    `json:"fileURL"`
	Articles    []OrderDetails
}

// ---------- Order details structure ----------
type OrderDetails struct {
	Article string          `json:"article"`
	Amount  decimal.Decimal `json:"amount"`
}

// Calendar's SynkTokens
type Calendar struct {
	CalendarID string
	SyncToken  string
}
