package store

import (
	"capitalbank/db"
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
)

func SaveState(data DataState) {

	// Connect to database
	// db, err := sql.Open("sqlserver", "server=bold;database=capital2010;Integrated Security=SSPI")
	// if err != nil {
	// 	fmt.Println("Error opening database:", err.Error())
	// }
	// defer db.Close()

	// Insert data into database
	_, err := db.DB.Exec("INSERT INTO _test_table (status, type, phase) VALUES (@p1, @p2, @p3)",
		sql.Named("p1", data.Status),
		sql.Named("p2", data.Type),
		sql.Named("p3", data.Phase))
	if err != nil {
		fmt.Println("Error inserting data into database:", err.Error())
	}
}
