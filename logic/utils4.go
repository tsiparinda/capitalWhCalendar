package logic

import (
	"capitalWhCalendar/store"
	"time"
)

func EarliestFBADTrue(balances *[]store.DataCheckBalance) (time.Time, error) {
	//var earliest *store.DataCheckBalance
	 earliestDate := time.Now().AddDate(0, 0, 2)

	for _, balance := range *balances {
		if balance.FBAD {
			if balance.Dpd.Before(earliestDate) {
				//earliest = &balance
				earliestDate = balance.Dpd
			}
		}
	}
	return earliestDate, nil
}
