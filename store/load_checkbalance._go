package store

import (
	"capitalbank/db"
	"capitalbank/logger"
	"database/sql"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// ...
// load from database and fill the Account.DataCheckBalance structure for check the early bad balance by accounts
func LoadCheckBalance(datefrom time.Time, acc *[]Account) error {
	for i, account := range *acc {
		//	fmt.Printf("LoadCheckBalance acc: %v, %v, %v, %v\n", datefrom, account.BankRegistr, sql.Named("p1", datefrom.Format("02.01.2006")), sql.Named("p2", account.BankRegistr))
		// Select data from database
		rows, err := db.DB.Query("SELECT DateBal,  cast(BankRegistr as varchar(50)), fBad FROM bank_checkbalance where DateBal >= @p1 and BankRegistr = @p2",
			sql.Named("p1", datefrom.Format("02.01.2006")),
			sql.Named("p2", account.BankRegistr))
		if err != nil {
			logger.Log.Info("Error loading accounts checkbalance from database:", err.Error())
			return err
		}
		defer rows.Close()

		//	fmt.Printf("rows: %v\n", rows)
		for rows.Next() {
			var a DataCheckBalance
			// Scan each column into the corresponding field of an Account. Adjust this line as needed based on your table structure.
			err = rows.Scan(&a.Dpd, &a.BankRegistr, &a.FBAD)
			if err != nil {
				logger.Log.Info("Error scanning accounts rows:", err.Error())
				return err
			}
			//	fmt.Printf("a: %v\n", a)
			(*acc)[i].BalanceState = append((*acc)[i].BalanceState, a)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			logger.Log.Info("Error iterating accounts rows:", err.Error())
			return err
		}
	}
	return nil
}
