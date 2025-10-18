# Application as a service to push and update events from local accounting system to Google Calendar

## Order → Google Calendar Synchronization Logic

### 1. Identify new orders
- Select all orders from the database where:
  - `ExpectedDate > CURRENT_DATE`
  - `whcal_EventID IS NULL`
- These represent orders that have not yet been pushed to Google Calendar.

### 2. Calendar list and sync state
- The system maintains a list of warehouse calendars in the database.
- Each calendar record stores its `calendarId` and the last known `syncToken`.
- The `syncToken` represents the last successful synchronization point for that calendar.

### 3. Synchronize existing events
- For each calendar:
  - Use the stored `syncToken` to call:
    ```
    events.list(calendarId, syncToken=…)
    ```
  - This returns only events that have been **created, updated, moved, or deleted** since the last sync.
  - If the `syncToken` is expired, restart the sync with a full `events.list` request using a `timeMin` filter.
- Match each returned event to the corresponding order using:
        ExtendedProperties[order_id]
- Update the local database with any detected changes (status, date/time, calendarId, etc.).

### 4. Insert new orders into the calendar
- For each new order (see step 1):
- Create a new Google Calendar event in the appropriate warehouse calendar.
- Store the resulting `eventId` in the order record (`whcal_EventID`).
- Optionally assign a special color (e.g., red) to visually mark newly created events.

## Compile and Install
- uncomment block 
	// ONE TIME ONLY!!!
	// auth.Manual_auth()
- go run .
- go buld .
- copy capitalWhCalendar.exe to \\bold\c$\Program files\capitalWhCalendar
- copy sevrets folder to \\bold\c$\Program files\capitalWhCalendar
- give a full permissions on folder \\bold\c$\Program files\capitalWhCalendar to account who will run .exe!!!