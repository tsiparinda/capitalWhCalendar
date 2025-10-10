package main

import (
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"
	"capitalWhCalendar/logic"
	"context"
	"fmt"
	"os"
)

func main() {
	ctx := context.Background()
	err := db.DB.Ping()
	if err != nil {
		fmt.Println("Database connection is not active")
		os.Exit(1)
	}

	fields := make(map[string]interface{})
	fields["logLevel"] = logger.Log.GetLevel()
	// Add more fields dynamically...
	// fields["location"] = "Earth"

	logger.Log.Info("Program was started")

	// Create service for Disk access
	driveSrv, err := logic.CreateDriveService()
	if err != nil {
		logger.Log.Fatal(err)
	}
	
	// Create service for Disk access
	calSrv, err := logic.CreateCalService(ctx)
	if err != nil {
		logger.Log.Fatal(err)
	}

	logic.SyncOrdersEvents(ctx, calSrv)
	logic.SendNewOrders(ctx, driveSrv, calSrv)

}
