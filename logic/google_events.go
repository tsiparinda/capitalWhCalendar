package logic

import (
	"capitalWhCalendar/logger"
	"capitalWhCalendar/store"
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func SendNewOrders() {
	ctx := context.Background()

	orders := []store.Order{}

	err := store.LoadOrders(&orders)
	if err != nil {
		logger.Log.Errorf("SendNewOrders: Error from LoadOrders:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"orders": orders,
	}).Trace("SendNewOrders: Orders loaded by LoadOrders")

	// Создание сервиса с сервисным аккаунтом
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		// log.Fatalf("StartExchangeEvents: Unable to create Calendar service: %v", err)
		logger.Log.Errorf("SendNewOrders: Unable to create Calendar service: %v", err.Error())
	}

	// cycle to send events
	for _, p := range orders {

		// Пример: создание события
		eventID, err := createEvent(ctx, srv, p.CalendarID, p.Summary, p.Description, p.OperID, p.Start, p.End, p.ColorId)
		if err != nil {
			logger.Log.Errorf("SendNewOrders: Error creating event: %v", err)
		}
		p.EventID = eventID
		// update order
		store.LinkOrder2Event(p)
		logger.Log.Debugf("SendNewOrders: Created event ID:", eventID)
	}

}

func SyncOrdersEvents() {
	ctx := context.Background()

	// Создание сервиса с сервисным аккаунтом
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		// log.Fatalf("StartExchangeEvents: Unable to create Calendar service: %v", err)
		logger.Log.Errorf("SyncOrdersEvents: Unable to create Calendar service: %v", err.Error())
	}

	calendars := []store.Calendar{}

	err = store.LoadCalendars(&calendars)
	if err != nil {
		logger.Log.Errorf("SyncOrdersEvents: Error from LoadCalendars:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"calendars": calendars,
	}).Trace("SyncOrdersEvents: Calendars loaded by LoadCalendars")
	// Список календарей и их syncToken
	// calendars := []store.Calendar{
	// 	{CalendarID: "9e36c3271e54afacec8d61b01e8db14cbd0a87cac5ee5ff74226f92305d16de4@group.calendar.google.com"},
	// 	{CalendarID: "3c902775b6f7349e63e32371db3b69989dfd8f5d600c85198705a4b9a55f75a5@group.calendar.google.com"},
	// }

	// Локальная база заказов
	orders := []store.Order{}

	// Синхронизируем каждый календарь
	for i, cal := range calendars {
		newToken, events, err := syncCalendar(ctx, srv, cal.CalendarID, cal.SyncToken)
		if err != nil {
			logger.Log.Errorf("SyncOrdersEvents:Error syncing calendar %s: %v", cal.CalendarID, err)
			continue
		}
		// Обновляем syncToken
		calendars[i].SyncToken = newToken

		// Обновляем локальную базу заказов
		for _, e := range events {
			operID := ""
			if e.ExtendedProperties != nil && e.ExtendedProperties.Private != nil {
				operID = e.ExtendedProperties.Private["operid"]
			}

			start, _ := parseEventDateTime(e.Start)
			end, _ := parseEventDateTime(e.End)

			if operID != "" {
				orders = append(orders, store.Order{
					OperID:      operID,
					Summary:     e.Summary,
					Start:       start,
					End:         end,
					Description: e.Description,
					CalendarID:  cal.CalendarID,
					EventID:     e.Id,
				})
			}
		}
	}

	//update calendars tokens
	err = store.UpdateCalendarTokens(calendars)
	if err != nil {
		logger.Log.Errorf("SyncOrdersEvents: Error from UpdateCalendarTokens:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"calendars": calendars,
	}).Trace("SyncOrdersEvents: Calendars updated by UpdateCalendarTokens")

	// Пример вывода локальной базы
	for _, o := range orders {
		logger.Log.Info("Order %s | %s | %s | %s\n", o.OperID, o.Summary, o.Summary, o.CalendarID, o.Start, o.End, o.EventID)
	}

}
