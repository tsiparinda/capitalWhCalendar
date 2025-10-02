package logic

import (
	"capitalWhCalendar/config"
	"capitalWhCalendar/logger"

	"capitalWhCalendar/store"
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/drive/v3"
)

func SendNewOrders(ctx context.Context, driveSrv *drive.Service, calSrv *calendar.Service) {

	orders := []store.Order{}

	// fill Order structure with details
	err := store.LoadOrders(&orders)
	if err != nil {
		logger.Log.Errorf("SendNewOrders: Error from LoadOrders:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"orders": orders,
	}).Trace("SendNewOrders: Orders loaded by LoadOrders")

	folderID := config.Config["folderID"].(string)

	var storedays int
	// take requestdays from config file and convert to int
	if value, ok := config.Config["fileStorageTimeDays"].(float64); ok {
		// The value is a float64, handle it accordingly
		storedays = int(value)
	} else {
		logger.Log.Info("SendNewOrders: Error loading fileStorageTimeDays from config:", err.Error())
		storedays = 7
	}

	// delete old files from Drive
	deleteOldFiles(driveSrv, folderID, storedays)

	// generate PDF for attachments
	if err := store.CreatePDF(driveSrv, &orders, folderID); err != nil {
		logger.Log.Errorf("SendNewOrders: Error from GeneratePDF:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"orders": orders,
	}).Trace("SendNewOrders: Orders loaded to pdf")

	// cycle to send events
	for _, p := range orders {

		// Create Event
		eventID, err := createEvent(ctx, calSrv, p.CalendarID, p.Summary, p.Description, p.OperID, p.ColorId, p.FileURL, p.Start, p.End)
		if err != nil {
			logger.Log.Errorf("SendNewOrders: Error creating event: %v", err)
		}
		p.EventID = eventID
		// Link Order with Event - save eventID in order's parameter
		store.LinkOrder2Event(p)
		logger.Log.Debugf("SendNewOrders: Created event ID:", eventID)
	}

}

func SyncOrdersEvents(ctx context.Context, calSrv *calendar.Service) {

	// declaration of local Orders DB
	orders := []store.Order{}
	calendars := []store.Calendar{}

	// Take a list of Calendars
	err := store.LoadCalendars(&calendars)
	if err != nil {
		logger.Log.Errorf("SyncOrdersEvents: Error from LoadCalendars:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"calendars": calendars,
	}).Trace("SyncOrdersEvents: Calendars loaded by LoadCalendars")

	// Syncronise each calendar by SyncToken (delta of changes)
	for i, cal := range calendars {
		newToken, events, err := syncCalendar(ctx, calSrv, cal.CalendarID, cal.SyncToken)
		if err != nil {
			logger.Log.Errorf("SyncOrdersEvents:Error syncing calendar %s: %v", cal.CalendarID, err)
			continue
		}
		// update syncToken
		calendars[i].SyncToken = newToken

		// fill in the local Orders DB
		for _, e := range events {
			operID := ""
			if e.ExtendedProperties != nil && e.ExtendedProperties.Private != nil {
				operID = e.ExtendedProperties.Private["operid"]
			}

			start, err := parseEventDateTime(e.Start)
			if err != nil {
				logger.Log.Warnf("Failed to parse start time for event %s: %v", e.Id, err)
				continue
			}
			end, err := parseEventDateTime(e.End)
			if err != nil {
				logger.Log.Warnf("Failed to parse start time for event %s: %v", e.Id, err)
				continue
			}
			fileURL := ""
			if e.Attachments != nil && len(e.Attachments) > 0 {
				for _, a := range e.Attachments {
					if a.FileUrl != "" {
						fileURL = a.FileUrl
						break
					}
				}
			}
			if operID != "" {
				orders = append(orders, store.Order{
					OperID:      operID,
					Summary:     e.Summary,
					Start:       start,
					End:         end,
					ColorId:     e.ColorId,
					FileURL:     fileURL,
					Description: e.Description,
					CalendarID:  cal.CalendarID,
					EventID:     e.Id,
				})
			}
		}
	}

	// update Order parameters in SQL DB
	err = store.OrdersUpdate(&orders)
	if err != nil {
		logger.Log.Errorf("SyncOrdersEvents: Error from OrdersUpdate:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"calendars": calendars,
	}).Trace("SyncOrdersEvents: Orders updated by OrdersUpdate")

	//update Calendars tokens in SQL DB
	err = store.UpdateCalendarTokens(&calendars)
	if err != nil {
		logger.Log.Errorf("SyncOrdersEvents: Error from UpdateCalendarTokens:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"calendars": calendars,
	}).Trace("SyncOrdersEvents: Calendars updated by UpdateCalendarTokens")
}
