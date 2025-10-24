package main

import (
	"capitalWhCalendar/auth"
	"capitalWhCalendar/db"
	"capitalWhCalendar/logger"
	"capitalWhCalendar/logic"
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
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

	// old auth block
	// Create service for Disk access
	// driveSrv, err := logic.CreateDriveService()
	// if err != nil {
	// 	logger.Log.Fatal(err)
	// }
	// // Create service for Disk access
	// calSrv, err := logic.CreateCalService(ctx)
	// if err != nil {
	// 	logger.Log.Fatal(err)
	// }

	//new auth

	// ONE TIME ONLY!!!
	// uncommen next line and exe: go run . - terminal will wait your answer
	// copy link from terminal to browser on machine where this application will work permanently
	// login to Google by account from where was client-service.json received
	// Continue and check all the permissions
	// after you will have error 404, copy the token = part of address from &code=... till next word
	// put this line to terminal
	// the file secrets/ google-token.json is ready to using - copy it to same folder on permanently working machine
	// auth.Manual_auth()

	client := auth.GetServiceClient()

	driveSrv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Drive service: %v", err)
	}

	calSrv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Calendar service: %v", err)
	}

	// logic
	logic.SyncOrdersEvents(ctx, calSrv)
	logic.SendNewOrders(ctx, driveSrv, calSrv)

}
