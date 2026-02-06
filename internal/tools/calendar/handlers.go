package calendar

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/calendar/v3"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- list_calendars ---

type ListCalendarsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
}

type ListCalendarsOutput struct {
	Calendars []CalendarSummary `json:"calendars"`
}

func createListCalendarsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListCalendarsInput, ListCalendarsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListCalendarsInput) (*mcp.CallToolResult, ListCalendarsOutput, error) {
		srv, err := factory.Calendar(ctx, input.UserEmail)
		if err != nil {
			return nil, ListCalendarsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.CalendarList.List().Context(ctx).Do()
		if err != nil {
			return nil, ListCalendarsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		calendars := make([]CalendarSummary, 0, len(result.Items))
		rb := response.New()
		rb.Header("Calendars")
		rb.KeyValue("Count", len(result.Items))
		rb.Blank()

		for _, c := range result.Items {
			cal := calendarToSummary(c)
			calendars = append(calendars, cal)

			primary := ""
			if cal.Primary {
				primary = " (primary)"
			}
			rb.Item("%s%s", cal.Summary, primary)
			rb.Line("    ID: %s", cal.ID)
			if cal.TimeZone != "" {
				rb.Line("    Timezone: %s", cal.TimeZone)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListCalendarsOutput{Calendars: calendars}, nil
	}
}

// --- get_events ---

type GetEventsInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	CalendarID string `json:"calendar_id,omitempty" jsonschema_description:"Calendar ID (default: primary)"`
	TimeMin    string `json:"time_min,omitempty" jsonschema_description:"Start of time range (RFC3339 e.g. 2025-06-15T00:00:00Z)"`
	TimeMax    string `json:"time_max,omitempty" jsonschema_description:"End of time range (RFC3339)"`
	MaxResults int    `json:"max_results,omitempty" jsonschema_description:"Maximum events to return (default 25)"`
	Query      string `json:"query,omitempty" jsonschema_description:"Free-text search within event fields"`
	EventID    string `json:"event_id,omitempty" jsonschema_description:"Specific event ID to retrieve (ignores time filters)"`
	Detailed   bool   `json:"detailed,omitempty" jsonschema_description:"Include full event details including attendees"`
}

type GetEventsOutput struct {
	Events []EventSummary `json:"events"`
}

func createGetEventsHandler(factory *services.Factory) mcp.ToolHandlerFor[GetEventsInput, GetEventsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEventsInput) (*mcp.CallToolResult, GetEventsOutput, error) {
		srv, err := factory.Calendar(ctx, input.UserEmail)
		if err != nil {
			return nil, GetEventsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		calID := input.CalendarID
		if calID == "" {
			calID = "primary"
		}

		// Single event retrieval
		if input.EventID != "" {
			event, err := srv.Events.Get(calID, input.EventID).Context(ctx).Do()
			if err != nil {
				return nil, GetEventsOutput{}, middleware.HandleGoogleAPIError(err)
			}

			es := eventToSummary(event)
			rb := response.New()
			rb.Header("Calendar Event")
			formatEventDetail(rb, es)

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
			}, GetEventsOutput{Events: []EventSummary{es}}, nil
		}

		// List events
		if input.MaxResults == 0 {
			input.MaxResults = 25
		}

		call := srv.Events.List(calID).
			MaxResults(int64(input.MaxResults)).
			SingleEvents(true).
			OrderBy("startTime").
			Context(ctx)

		if input.TimeMin != "" {
			call = call.TimeMin(input.TimeMin)
		}
		if input.TimeMax != "" {
			call = call.TimeMax(input.TimeMax)
		}
		if input.Query != "" {
			call = call.Q(input.Query)
		}

		result, err := call.Do()
		if err != nil {
			return nil, GetEventsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		events := make([]EventSummary, 0, len(result.Items))
		rb := response.New()
		rb.Header("Calendar Events")
		rb.KeyValue("Calendar", calID)
		rb.KeyValue("Events", len(result.Items))
		rb.Blank()

		for _, e := range result.Items {
			es := eventToSummary(e)
			events = append(events, es)

			if input.Detailed {
				formatEventDetail(rb, es)
				rb.Separator()
			} else {
				rb.Item("%s", es.Summary)
				rb.Line("    %s → %s", es.Start, es.End)
				if es.Location != "" {
					rb.Line("    Location: %s", es.Location)
				}
				rb.Line("    ID: %s", es.ID)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetEventsOutput{Events: events}, nil
	}
}

// --- create_event ---

type CreateEventInput struct {
	UserEmail   string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Summary     string   `json:"summary" jsonschema:"required" jsonschema_description:"Event title"`
	StartTime   string   `json:"start_time" jsonschema:"required" jsonschema_description:"Start time (RFC3339 or date for all-day)"`
	EndTime     string   `json:"end_time" jsonschema:"required" jsonschema_description:"End time (RFC3339 or date for all-day)"`
	CalendarID  string   `json:"calendar_id,omitempty" jsonschema_description:"Calendar ID (default: primary)"`
	Description string   `json:"description,omitempty" jsonschema_description:"Event description"`
	Location    string   `json:"location,omitempty" jsonschema_description:"Event location"`
	Attendees   []string `json:"attendees,omitempty" jsonschema_description:"Attendee email addresses"`
	Timezone    string   `json:"timezone,omitempty" jsonschema_description:"Timezone (e.g. America/New_York)"`
	Reminders   string   `json:"reminders,omitempty" jsonschema_description:"JSON array of reminders [{method: popup/email, minutes: N}]"`
	AddMeet     bool     `json:"add_google_meet,omitempty" jsonschema_description:"Add a Google Meet video conference"`
}

func createCreateEventHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateEventInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateEventInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Calendar(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		calID := input.CalendarID
		if calID == "" {
			calID = "primary"
		}

		event := &calendar.Event{
			Summary:     input.Summary,
			Description: input.Description,
			Location:    input.Location,
			Attendees:   buildAttendees(input.Attendees),
		}

		// Set start/end times
		event.Start = buildEventDateTime(input.StartTime, input.Timezone)
		event.End = buildEventDateTime(input.EndTime, input.Timezone)

		// Reminders
		if input.Reminders != "" {
			reminders, err := parseReminders(input.Reminders)
			if err != nil {
				return nil, nil, err
			}
			event.Reminders = &calendar.EventReminders{
				UseDefault: false,
				Overrides:  reminders,
			}
		}

		// Google Meet
		if input.AddMeet {
			event.ConferenceData = &calendar.ConferenceData{
				CreateRequest: &calendar.CreateConferenceRequest{
					RequestId: fmt.Sprintf("meet-%s", input.Summary),
				},
			}
		}

		call := srv.Events.Insert(calID, event).Context(ctx)
		if input.AddMeet {
			call = call.ConferenceDataVersion(1)
		}

		created, err := call.Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Event Created")
		rb.KeyValue("Summary", created.Summary)
		rb.KeyValue("Start", formatEventTime(created.Start))
		rb.KeyValue("End", formatEventTime(created.End))
		rb.KeyValue("ID", created.Id)
		if created.HtmlLink != "" {
			rb.KeyValue("Link", created.HtmlLink)
		}
		if created.ConferenceData != nil && len(created.ConferenceData.EntryPoints) > 0 {
			for _, ep := range created.ConferenceData.EntryPoints {
				if ep.EntryPointType == "video" {
					rb.KeyValue("Google Meet", ep.Uri)
				}
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- modify_event ---

type ModifyEventInput struct {
	UserEmail   string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	EventID     string   `json:"event_id" jsonschema:"required" jsonschema_description:"The ID of the event to modify"`
	CalendarID  string   `json:"calendar_id,omitempty" jsonschema_description:"Calendar ID (default: primary)"`
	Summary     string   `json:"summary,omitempty" jsonschema_description:"New event title"`
	StartTime   string   `json:"start_time,omitempty" jsonschema_description:"New start time (RFC3339)"`
	EndTime     string   `json:"end_time,omitempty" jsonschema_description:"New end time (RFC3339)"`
	Description string   `json:"description,omitempty" jsonschema_description:"New event description"`
	Location    string   `json:"location,omitempty" jsonschema_description:"New event location"`
	Attendees   []string `json:"attendees,omitempty" jsonschema_description:"Updated attendee email list (replaces existing)"`
	Timezone    string   `json:"timezone,omitempty" jsonschema_description:"New timezone"`
}

func createModifyEventHandler(factory *services.Factory) mcp.ToolHandlerFor[ModifyEventInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ModifyEventInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Calendar(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		calID := input.CalendarID
		if calID == "" {
			calID = "primary"
		}

		// Get existing event
		existing, err := srv.Events.Get(calID, input.EventID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// Apply updates
		if input.Summary != "" {
			existing.Summary = input.Summary
		}
		if input.Description != "" {
			existing.Description = input.Description
		}
		if input.Location != "" {
			existing.Location = input.Location
		}
		if input.StartTime != "" {
			existing.Start = buildEventDateTime(input.StartTime, input.Timezone)
		}
		if input.EndTime != "" {
			existing.End = buildEventDateTime(input.EndTime, input.Timezone)
		}
		if input.Attendees != nil {
			existing.Attendees = buildAttendees(input.Attendees)
		}

		updated, err := srv.Events.Update(calID, input.EventID, existing).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Event Modified")
		rb.KeyValue("Summary", updated.Summary)
		rb.KeyValue("Start", formatEventTime(updated.Start))
		rb.KeyValue("End", formatEventTime(updated.End))
		rb.KeyValue("ID", updated.Id)
		if updated.HtmlLink != "" {
			rb.KeyValue("Link", updated.HtmlLink)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- delete_event ---

type DeleteEventInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	EventID    string `json:"event_id" jsonschema:"required" jsonschema_description:"The ID of the event to delete"`
	CalendarID string `json:"calendar_id,omitempty" jsonschema_description:"Calendar ID (default: primary)"`
}

func createDeleteEventHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteEventInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteEventInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Calendar(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		calID := input.CalendarID
		if calID == "" {
			calID = "primary"
		}

		err = srv.Events.Delete(calID, input.EventID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Event Deleted")
		rb.KeyValue("Event ID", input.EventID)
		rb.KeyValue("Calendar", calID)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- query_freebusy (extended) ---

type QueryFreeBusyInput struct {
	UserEmail  string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TimeMin    string   `json:"time_min" jsonschema:"required" jsonschema_description:"Start of time range (RFC3339)"`
	TimeMax    string   `json:"time_max" jsonschema:"required" jsonschema_description:"End of time range (RFC3339)"`
	CalendarIDs []string `json:"calendar_ids,omitempty" jsonschema_description:"Calendar IDs to check (default: primary)"`
}

type FreeBusyPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type FreeBusyCalendarResult struct {
	CalendarID string           `json:"calendar_id"`
	Busy       []FreeBusyPeriod `json:"busy"`
}

type QueryFreeBusyOutput struct {
	Calendars []FreeBusyCalendarResult `json:"calendars"`
}

func createQueryFreeBusyHandler(factory *services.Factory) mcp.ToolHandlerFor[QueryFreeBusyInput, QueryFreeBusyOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input QueryFreeBusyInput) (*mcp.CallToolResult, QueryFreeBusyOutput, error) {
		srv, err := factory.Calendar(ctx, input.UserEmail)
		if err != nil {
			return nil, QueryFreeBusyOutput{}, middleware.HandleGoogleAPIError(err)
		}

		calIDs := input.CalendarIDs
		if len(calIDs) == 0 {
			calIDs = []string{"primary"}
		}

		items := make([]*calendar.FreeBusyRequestItem, 0, len(calIDs))
		for _, id := range calIDs {
			items = append(items, &calendar.FreeBusyRequestItem{Id: id})
		}

		result, err := srv.Freebusy.Query(&calendar.FreeBusyRequest{
			TimeMin: input.TimeMin,
			TimeMax: input.TimeMax,
			Items:   items,
		}).Context(ctx).Do()
		if err != nil {
			return nil, QueryFreeBusyOutput{}, middleware.HandleGoogleAPIError(err)
		}

		calendars := make([]FreeBusyCalendarResult, 0, len(result.Calendars))
		rb := response.New()
		rb.Header("Free/Busy Results")
		rb.KeyValue("Time Range", fmt.Sprintf("%s → %s", input.TimeMin, input.TimeMax))
		rb.Blank()

		for calID, cal := range result.Calendars {
			periods := make([]FreeBusyPeriod, 0, len(cal.Busy))
			rb.Section("Calendar: %s", calID)

			if len(cal.Busy) == 0 {
				rb.Item("Free (no busy periods)")
			}
			for _, b := range cal.Busy {
				periods = append(periods, FreeBusyPeriod{Start: b.Start, End: b.End})
				rb.Item("Busy: %s → %s", b.Start, b.End)
			}

			calendars = append(calendars, FreeBusyCalendarResult{
				CalendarID: calID,
				Busy:       periods,
			})
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, QueryFreeBusyOutput{Calendars: calendars}, nil
	}
}

// --- helper functions ---

// buildEventDateTime creates an EventDateTime from a time string.
func buildEventDateTime(timeStr, timezone string) *calendar.EventDateTime {
	// Check if it's an all-day event (date only, no T or time component)
	if len(timeStr) <= 10 {
		return &calendar.EventDateTime{Date: timeStr}
	}
	edt := &calendar.EventDateTime{DateTime: timeStr}
	if timezone != "" {
		edt.TimeZone = timezone
	}
	return edt
}

// formatEventDetail writes detailed event info to the response builder.
func formatEventDetail(rb *response.Builder, es EventSummary) {
	rb.KeyValue("Summary", es.Summary)
	rb.KeyValue("Start", es.Start)
	rb.KeyValue("End", es.End)
	if es.Location != "" {
		rb.KeyValue("Location", es.Location)
	}
	if es.Description != "" {
		rb.KeyValue("Description", es.Description)
	}
	if es.Organizer != "" {
		rb.KeyValue("Organizer", es.Organizer)
	}
	if len(es.Attendees) > 0 {
		rb.Section("Attendees")
		for _, a := range es.Attendees {
			rb.Item("%s", a)
		}
	}
	rb.KeyValue("Status", es.Status)
	rb.KeyValue("ID", es.ID)
	if es.HTMLLink != "" {
		rb.KeyValue("Link", es.HTMLLink)
	}
}
