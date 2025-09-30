package store

import (
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"
	"database/sql"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

func LoadCalendars(calendars *[]Calendar) error {

	// Select data from database
	rows, err := db.DB.Query("SELECT CalendarID, SyncToken  FROM whcal_CalendarsList")
	if err != nil {
		logger.Log.Errorf("LoadCalendars: Error loading calendars from database:", err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p Calendar
		// Scan each column into the corresponding field of an Account. Adjust this line as needed based on your table structure.
		err = rows.Scan(&p.CalendarID, &p.SyncToken)
		if err != nil {
			logger.Log.Errorf("LoadCalendars: Error scanning calendars rows:", err.Error())
			return err
		}

		*calendars = append(*calendars, p)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		logger.Log.Errorf("LoadCalendars: Error iterating calendars rows:", err.Error())
		return err
	}

	return nil
}

func UpdateCalendarTokens(cal []Calendar) error {
	for _, c := range cal {
		// Update data of order in DB
		if _, err := db.DB.Exec("exec whcal_CalendarUpdate @CalendarID, @SyncToken",
			sql.Named("CalendarID", c.CalendarID),
			sql.Named("SyncToken", c.SyncToken)); err != nil {
			logger.Log.WithFields(logrus.Fields{
				"CalendarID": c.CalendarID,
			}).Debugf("UpdateCalendarTokens: Error run sp.[whcal_CalendarUpdate]: ", err.Error())
			return err
		}
	}
	return nil
}
