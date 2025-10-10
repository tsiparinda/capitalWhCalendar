package logic

import (
	"capitalWhCalendar/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"context"
	"os"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
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

// clean old files from the work folder on Drive
func deleteOldFiles(driveSrv *drive.Service, folderID string, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)

	q := fmt.Sprintf("'%s' in parents and mimeType='application/pdf' and createdTime<'%s'", folderID, cutoff)
	pageToken := ""
	for {
		resp, err := driveSrv.Files.List().Q(q).Fields("nextPageToken, files(id, name)").PageToken(pageToken).Do()
		if err != nil {
			return fmt.Errorf("deleteOldFiles: %v", err)
		}

		for _, f := range resp.Files {
			err := driveSrv.Files.Delete(f.Id).Do()
			if err != nil {
				logger.Log.Errorf("deleteOldFiles: Failed to delete file %s: %v", f.Name, err)
			} else {
				logger.Log.Infof("deleteOldFiles: Deleted old file %s", f.Name)
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return nil
}

// create Event
func createEvent(ctx context.Context, srv *calendar.Service, calendarID, summary, description, operid, colorid, fileURL string, start, end time.Time) (string, error) {

	event := &calendar.Event{

		Summary:     summary,
		Description: description,
		//ColorId:     "7",
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "",
		},
		End: &calendar.EventDateTime{
			//DateTime: end.Format("2006-01-02T15:15:05"),
			DateTime: end.Format(time.RFC3339),
			TimeZone: "",
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"operid":    operid,
				"warehouse": "new",
				"status":    "pending", // например "pending"
			},
		},
		ColorId: colorid,
	}

	// ✅ Add Attachment only if link is existed
	if fileURL != "" {
		event.Attachments = []*calendar.EventAttachment{
			{
				FileUrl:  fileURL,
				Title:    "Заказ #" + operid,
				MimeType: "application/pdf",
			},
		}
	}

	createdEvent, err := srv.Events.Insert(calendarID, event).
		SupportsAttachments(true). // <-- обязательно!
		Do()

	if err != nil {
		return "", err
	}
	
	return createdEvent.Id, nil
}

// get Event by ID
func getEvent(ctx context.Context, srv *calendar.Service, calendarID, eventID string) (*calendar.Event, error) {
	return srv.Events.Get(calendarID, eventID).Do()
}

// update Event
func updateEvent(ctx context.Context, srv *calendar.Service, calendarID string, event *calendar.Event) error {
	_, err := srv.Events.Update(calendarID, event.Id, event).Do()
	return err
}

// delete Event
func deleteEvent(ctx context.Context, srv *calendar.Service, calendarID, eventID string) error {
	return srv.Events.Delete(calendarID, eventID).Do()
}

// syncCalendar - get changes by syncToken and take a new synctoken
func syncCalendar(ctx context.Context, srv *calendar.Service, calendarID, syncToken string) (string, []*calendar.Event, error) {
	call := srv.Events.List(calendarID).
		ShowDeleted(true).
		SingleEvents(true)

	if syncToken != "" {
		call = call.SyncToken(syncToken)
	}

	events, err := call.Do()
	if err != nil {
		// Если syncToken устарел, делаем полный обход
		if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 410 {
			fmt.Println("Sync token expired, doing full sync")
			return syncCalendar(ctx, srv, calendarID, "")
		}
		return "", nil, err
	}

	return events.NextSyncToken, events.Items, nil
}

func parseEventDateTime(edt *calendar.EventDateTime) (time.Time, error) {
	if edt == nil {
		return time.Time{}, fmt.Errorf("empty EventDateTime")
	}
	if edt.DateTime != "" {
		// обычное событие
		return time.Parse(time.RFC3339, edt.DateTime)
	}
	if edt.Date != "" {
		// событие на весь день (без времени, берём начало дня)
		return time.Parse("2006-01-02", edt.Date)
	}
	return time.Time{}, fmt.Errorf("no date found in EventDateTime")
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

// read token
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// save token
func saveToken(path string, token *oauth2.Token) {
	f, _ := os.Create(path)
	defer f.Close()
	json.NewEncoder(f).Encode(token)
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
