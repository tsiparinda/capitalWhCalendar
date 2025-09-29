package logic

import (
	"capitalWhCalendar/logger"
	"capitalWhCalendar/store"
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func StartExchangeEvents() {
	ctx := context.Background()

	orders := []store.Order{}

	err := store.LoadOrders(&orders)
	if err != nil {
		logger.Log.Errorf("StartExchangeEvents: Error from LoadOrders:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"orders": orders,
	}).Trace("StartExchangeEvents: Orders loaded by LoadOrders")

	// Создание сервиса с сервисным аккаунтом
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		// log.Fatalf("StartExchangeEvents: Unable to create Calendar service: %v", err)
		logger.Log.Errorf("StartExchangeEvents: Unable to create Calendar service: %v", err.Error())
	}

	// cycle to send events
	for _, p := range orders {

		// Пример: создание события
		eventID, err := createEvent(ctx, srv, p.CalendarID, p.Summary, p.Description, p.OperID, p.Start, p.End)
		if err != nil {
			logger.Log.Errorf("StartExchangeEvents: Error creating event: %v", err)
		}
		p.EventID = eventID
		// update order
		store.LinkOrder2Event(p)
		logger.Log.Debugf("StartExchangeEvents: Created event ID:", eventID)
	}

}
