package store

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	CalendarID  string    `json:"calendarID"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	OperID      string    `json:"operID"`
	EventID     string    `json:"evendID"`
	ColorId     string    `json:"colorid"`
	FileURL     string    `json:"fileURL"`
	Articles    []OrderDetails
}

// ---------- Структура заказа ----------
type OrderDetails struct {
	Article string
	Amount  decimal.Decimal
}

// Структура для хранения syncToken по каждому календарю
type Calendar struct {
	CalendarID string
	SyncToken  string
}
