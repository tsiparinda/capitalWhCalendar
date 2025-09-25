package utils

import (
	"time"
)

func GetShortDate(t time.Time) (time.Time, error) {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return t, nil
}
