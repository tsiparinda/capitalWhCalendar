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

	payments := []store.Payment{}

	// send privat
	err := store.LoadPaymentsPrivat(&payments)
	if err != nil {
		logger.Log.Info("StartExchangePayments: Error from LoadPaymentsPrivat:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"payments": payments,
	}).Trace("StartExchangePayments: Payment loaded by LoadPaymentsPrivat")

	// Создание сервиса с сервисным аккаунтом
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		log.Fatalf("Unable to create Calendar service: %v", err)
	}

	calendarID := "primary" // или ID нужного календаря

	// Пример: создание события
	eventID, err := createEvent(ctx, srv, calendarID, "Тестовое событие", "Описание события", time.Now(), time.Now().Add(1*time.Hour))
	if err != nil {
		log.Fatalf("Error creating event: %v", err)
	}
	fmt.Println("Created event ID:", eventID)

	// Пример: получение события
	ev, err := getEvent(ctx, srv, calendarID, eventID)
	if err != nil {
		log.Fatalf("Error getting event: %v", err)
	}
	fmt.Println("Event summary:", ev.Summary)

	// Пример: обновление события
	ev.Summary = "Обновлённое событие"
	if err := updateEvent(ctx, srv, calendarID, ev); err != nil {
		log.Fatalf("Error updating event: %v", err)
	}
	fmt.Println("Event updated")

	// Пример: удаление события
	if err := deleteEvent(ctx, srv, calendarID, eventID); err != nil {
		log.Fatalf("Error deleting event: %v", err)
	}
	fmt.Println("Event deleted")
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
