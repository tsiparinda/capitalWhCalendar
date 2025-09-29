package store

import (
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"
	"database/sql"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

// ...

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
