package calendar

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/api/calendar/v3"
)

// CalendarSummary is a compact representation of a Google Calendar.
type CalendarSummary struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
	TimeZone    string `json:"time_zone,omitempty"`
}

// EventSummary is a compact representation of a calendar event.
type EventSummary struct {
	ID          string   `json:"id"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Status      string   `json:"status,omitempty"`
	HTMLLink    string   `json:"html_link,omitempty"`
	Attendees   []string `json:"attendees,omitempty"`
	Organizer   string   `json:"organizer,omitempty"`
}

// calendarToSummary converts a CalendarListEntry to a compact summary.
func calendarToSummary(c *calendar.CalendarListEntry) CalendarSummary {
	return CalendarSummary{
		ID:          c.Id,
		Summary:     c.Summary,
		Description: c.Description,
		Primary:     c.Primary,
		TimeZone:    c.TimeZone,
	}
}

// eventToSummary converts a calendar Event to a compact summary.
func eventToSummary(e *calendar.Event) EventSummary {
	start := formatEventTime(e.Start)
	end := formatEventTime(e.End)

	attendees := make([]string, 0, len(e.Attendees))
	for _, a := range e.Attendees {
		attendees = append(attendees, formatAttendee(a))
	}

	var organizer string
	if e.Organizer != nil {
		organizer = e.Organizer.Email
		if e.Organizer.DisplayName != "" {
			organizer = fmt.Sprintf("%s (%s)", e.Organizer.DisplayName, e.Organizer.Email)
		}
	}

	return EventSummary{
		ID:          e.Id,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		Start:       start,
		End:         end,
		Status:      e.Status,
		HTMLLink:    e.HtmlLink,
		Attendees:   attendees,
		Organizer:   organizer,
	}
}

// formatEventTime returns a human-readable event time string.
func formatEventTime(et *calendar.EventDateTime) string {
	if et == nil {
		return ""
	}
	if et.Date != "" {
		return et.Date // All-day event
	}
	return et.DateTime
}

// formatAttendee returns a human-readable attendee string.
func formatAttendee(a *calendar.EventAttendee) string {
	parts := []string{a.Email}
	if a.DisplayName != "" {
		parts = []string{fmt.Sprintf("%s (%s)", a.DisplayName, a.Email)}
	}
	if a.ResponseStatus != "" {
		parts = append(parts, fmt.Sprintf("[%s]", a.ResponseStatus))
	}
	if a.Organizer {
		parts = append(parts, "(organizer)")
	}
	if a.Optional {
		parts = append(parts, "(optional)")
	}
	return strings.Join(parts, " ")
}

// ReminderSpec represents a reminder configuration.
type ReminderSpec struct {
	Method  string `json:"method"`
	Minutes int    `json:"minutes"`
}

// parseReminders parses a JSON string or list of reminder specs.
func parseReminders(input string) ([]*calendar.EventReminder, error) {
	if input == "" {
		return nil, nil
	}

	var specs []ReminderSpec
	if err := json.Unmarshal([]byte(input), &specs); err != nil {
		return nil, fmt.Errorf("parsing reminders JSON — expected [{\"method\":\"popup\",\"minutes\":15}]: %w", err)
	}

	reminders := make([]*calendar.EventReminder, 0, len(specs))
	for _, s := range specs {
		if s.Method != "popup" && s.Method != "email" {
			return nil, fmt.Errorf("invalid reminder method %q — use 'popup' or 'email'", s.Method)
		}
		if s.Minutes < 0 || s.Minutes > 40320 {
			return nil, fmt.Errorf("reminder minutes must be 0-40320, got %d", s.Minutes)
		}
		reminders = append(reminders, &calendar.EventReminder{
			Method:  s.Method,
			Minutes: int64(s.Minutes),
		})
	}

	return reminders, nil
}

// buildAttendees converts a list of email strings to calendar Attendees.
func buildAttendees(emails []string) []*calendar.EventAttendee {
	if len(emails) == 0 {
		return nil
	}
	attendees := make([]*calendar.EventAttendee, 0, len(emails))
	for _, email := range emails {
		attendees = append(attendees, &calendar.EventAttendee{Email: email})
	}
	return attendees
}
