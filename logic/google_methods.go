package logic

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

// Создание события
func createEvent(ctx context.Context, srv *calendar.Service, calendarID, summary, description, operid string, start, end time.Time, colorid string) (string, error) {

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
		ColorId: colorid,
	}
	createdEvent, err := srv.Events.Insert(calendarID, event).Do()
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
