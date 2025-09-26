package store

import (
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func LoadOrders(orders *[]Order) error {

	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		log.Fatal(err)
	}

	// Select data from database
	rows, err := db.DB.Query("SELECT calendarID,	summary,	description, [start], [end], OperID  FROM whcal_orders2send")
	if err != nil {
		logger.Log.Info("Error loading orders from database:", err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p Order
		// Scan each column into the corresponding field of an Account. Adjust this line as needed based on your table structure.
		err = rows.Scan(&p.CalendarID, &p.Summary, &p.Description, &p.Start, &p.End, &p.ID_Операции)
		if err != nil {
			logger.Log.Info("Error scanning orders rows:", err.Error())
			return err
		}
		// Приводим к нужной зоне
		p.Start = p.Start.In(loc)
		p.End = p.End.In(loc)

		*orders = append(*orders, p)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		logger.Log.Info("Error iterating orders rows:", err.Error())
		return err
	}

	return nil
}
