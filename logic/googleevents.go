package logic

import (
	"capitalWhCalendar/logger"
	"capitalWhCalendar/store"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func StartExchangeEvents() {
	ctx := context.Background()

	orders := []store.Order{}

	err := store.LoadOrders(&orders)
	if err != nil {
		logger.Log.Info("StartExchangeEvents: Error from LoadOrders:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"orders": orders,
	}).Trace("StartExchangeEvents: Payment loaded by LoadOrders")

	// Создание сервиса с сервисным аккаунтом
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		log.Fatalf("StartExchangeEvents: Unable to create Calendar service: %v", err)
	}

	// cycle
	for _, p := range orders {
		// Пример: создание события
		eventID, err := createEvent(ctx, srv, p.CalendarID, p.Summary, p.Description, p.Start, p.End)
		if err != nil {
			log.Fatalf("StartExchangeEvents: Error creating event: %v", err)
		}
		fmt.Println("StartExchangeEvents: Created event ID:", eventID)
	}
	////////////////
	//calendarID := "primary" // или ID нужного календаря

	// Пример: создание события
	// eventID, err := createEvent(ctx, srv, calendarID, "Тестовое событие", "Описание события", time.Now(), time.Now().Add(1*time.Hour))
	// if err != nil {
	// 	log.Fatalf("StartExchangeEvents: Error creating event: %v", err)
	// }
	// fmt.Println("StartExchangeEvents: Created event ID:", eventID)

	// Пример: получение события
	// ev, err := getEvent(ctx, srv, calendarID, eventID)
	// if err != nil {
	// 	log.Fatalf("StartExchangeEvents: Error getting event: %v", err)
	// }
	// fmt.Println("StartExchangeEvents: Event summary:", ev.Summary)

	// // Пример: обновление события
	// ev.Summary = "Обновлённое событие"
	// if err := updateEvent(ctx, srv, calendarID, ev); err != nil {
	// 	log.Fatalf("StartExchangeEvents: Error updating event: %v", err)
	// }
	// fmt.Println("StartExchangeEvents: Event updated")

	// // Пример: удаление события
	// if err := deleteEvent(ctx, srv, calendarID, eventID); err != nil {
	// 	log.Fatalf("Error deleting event: %v", err)
	// }
	// fmt.Println("Event deleted")
}

// Создание события
func createEvent(ctx context.Context, srv *calendar.Service, calendarID, summary, description string, start, end time.Time) (string, error) {
	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "Europe/Vienna",
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: "Europe/Vienna",
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
