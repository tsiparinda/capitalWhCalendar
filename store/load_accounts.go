package store

import (
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"

	_ "github.com/denisenkom/go-mssqldb"
)

func LoadAccounts(acc *[]Account) error {

	// Select data from database
	rows, err := db.DB.Query("SELECT Direction, Account, Bank, Token, cast(BankRegistr as varchar(50)) FROM bank_accounts_select")
	if err != nil {
		logger.Log.Info("Error loading accounts from database:", err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var a Account
		// Scan each column into the corresponding field of an Account. Adjust this line as needed based on your table structure.
		err = rows.Scan(&a.Direction, &a.Account, &a.Bank, &a.Token, &a.BankRegistr)
		if err != nil {
			logger.Log.Info("Error scanning accounts rows:", err.Error())
			return err
		}
		*acc = append(*acc, a)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		logger.Log.Info("Error iterating accounts rows:", err.Error())
		return err
	}

	return nil
}
