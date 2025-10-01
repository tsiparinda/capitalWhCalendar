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

// Функция очистки старых файлов на Drive
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

// Создание события
func createEvent(ctx context.Context, srv *calendar.Service, calendarID, summary, description, operid, colorid, fileURL string, start, end time.Time) (string, error) {

	event := &calendar.Event{

		Summary:     summary,
		Description: description,
		//ColorId:     "7",
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "Europe/Kiev",
		},
		End: &calendar.EventDateTime{
			//DateTime: end.Format("2006-01-02T15:15:05"),
			DateTime: end.Format(time.RFC3339),
			TimeZone: "Europe/Kiev",
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"operid":    operid,
				"warehouse": "new",
				"status":    "pending", // например "pending"
			},
		},
		Attachments: []*calendar.EventAttachment{
			{
				FileUrl:  fileURL,
				Title:    "Заказ #" + operid,
				MimeType: "application/pdf",
			},
		},
		ColorId: colorid,
	}

	createdEvent, err := srv.Events.Insert(calendarID, event).
		SupportsAttachments(true). // <-- обязательно!
		Do()

	if err != nil {
		return "", err
	}
	return createdEvent.Id, nil
}

// Получение события по ID
func getEvent(ctx context.Context, srv *calendar.Service, calendarID, eventID string) (*calendar.Event, error) {
	return srv.Events.Get(calendarID, eventID).Do()
}

// Обновление события
func updateEvent(ctx context.Context, srv *calendar.Service, calendarID string, event *calendar.Event) error {
	_, err := srv.Events.Update(calendarID, event.Id, event).Do()
	return err
}

// Удаление события
func deleteEvent(ctx context.Context, srv *calendar.Service, calendarID, eventID string) error {
	return srv.Events.Delete(calendarID, eventID).Do()
}

// syncCalendar получает изменения календаря по syncToken
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

func createDriveService() (*drive.Service, error) {
	//Читаем OAuth credentials
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client_secret.json: %v", err)
	}

	// Настраиваем конфиг
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
	}

	client := getClient(config)

	// Создаем сервис Drive
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %v", err)
	}
	return srv, nil
}

// Получаем OAuth клиент
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Чтение токена из файла
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

// Сохранение токена
func saveToken(path string, token *oauth2.Token) {
	f, _ := os.Create(path)
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Получение токена через браузер
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
