package logic

import (
	"context"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Создание события
func createEvent(ctx context.Context, srv *calendar.Service, calendarID, summary, description, operid string, start, end time.Time) (string, error) {

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
				"warehouse": "warehouse",
				"status":    "pending", // например "pending"
			},
		},
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
