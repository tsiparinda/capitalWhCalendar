package auth

import (
	"capitalWhCalendar/logger"
	"fmt"
	"net/http"

	"context"
	"os"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// create Drive service
func CreateDriveService() (*drive.Service, error) {
	// read OAuth credentials
	b, err := os.ReadFile("secrets/client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("CreateDriveService: unable to read client_secret.json: %v", err)
	}

	// get config
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("CreateDriveService: unable to parse client secret file: %v", err)
	}

	client := getClient(config)

	// createс Drive service
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("CreateDriveService: unable to create Drive service: %v", err)
	}
	return srv, nil
}

func CreateCalService(ctx context.Context) (*calendar.Service, error) {

	// Create service for calendar access
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("secrets/service-account.json"))
	if err != nil {
		// log.Fatalf("StartExchangeEvents: Unable to create Calendar service: %v", err)
		logger.Log.Errorf("CreateCalService: Unable to create Calendar service: %v", err.Error())
		return nil, err
	}
	return srv, nil
}

// Get OAuth client
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "secrets/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Get toket from browser
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open this URL in your browser:\n%v\n", authURL)
	fmt.Print("Enter the authorization code: ")
	var code string
	fmt.Scan(&code)
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		panic(err)
	}
	return tok
}
