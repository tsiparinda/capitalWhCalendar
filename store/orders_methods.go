package store

import (
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"
	"database/sql"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

func LoadOrders(orders *[]Order) error {

	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		log.Fatal(err)
	}

	// Select data from database
	rows, err := db.DB.Query("SELECT calendarID,	summary,	description, [start], [end], OperID, ColorID  FROM whcal_orders2send")
	if err != nil {
		logger.Log.Errorf("LoadOrders: Error loading orders from database:", err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p Order
		// Scan each column into the corresponding field of an Account. Adjust this line as needed based on your table structure.
		err = rows.Scan(&p.CalendarID, &p.Summary, &p.Description, &p.Start, &p.End, &p.OperID, &p.ColorId)
		if err != nil {
			logger.Log.Errorf("LoadOrders: Error scanning orders rows:", err.Error())
			return err
		}
		// Приводим к нужной зоне
		p.Start = p.Start.In(loc)
		p.End = p.End.In(loc)

		*orders = append(*orders, p)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		logger.Log.Errorf("LoadOrders: Error iterating orders rows:", err.Error())
		return err
	}

	// Check for errors from iterating over rows.
	if err := insertProduct(orders); err != nil {
		logger.Log.Errorf("LoadOrders: Error inserting orders rows:", err.Error())
		return err
	}

	return nil
}

// don't used
func LinkOrder2Event(p Order) {

	// Update data of order in DB
	if _, err := db.DB.Exec("exec whcal_LinkOrder2Event @OperID, @EventID",
		sql.Named("OperID", p.OperID),
		sql.Named("EventID", p.EventID)); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"OperID": p.OperID,
		}).Debugf("LinkOrder2Event: Error run sp.[whcal_LinkOrder2Event]: ", err.Error())
	}

}

func insertProduct(orders *[]Order) error {
	// in orders we have only new orders!!!
	// Select data from database

	for i := range *orders {
		o := &(*orders)[i]

		rows, err := db.DB.Query("exec whcal_GetOrderLines @OperID=@p1",
			sql.Named("p1", o.OperID))
		if err != nil {
			logger.Log.Errorf("insertProduct: Error loading orders lines from database:", err.Error())
			return err
		}

		for rows.Next() {
			var p OrderDetails
			// Scan each column into the corresponding field of an Account. Adjust this line as needed based on your table structure.
			err = rows.Scan(&p.Article, &p.Amount)
			if err != nil {
				logger.Log.Errorf("insertProduct: Error scanning order lines rows:", err.Error())
				return err
			}
			o.Articles = append(o.Articles, p)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			logger.Log.Errorf("LoadOrders: Error iterating orders rows:", err.Error())
			return err
		}
		defer rows.Close()
	}

	return nil
}
