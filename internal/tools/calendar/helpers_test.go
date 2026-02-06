package calendar

import (
	"testing"

	gcal "google.golang.org/api/calendar/v3"
)

func TestFormatEventTime(t *testing.T) {
	tests := []struct {
		name string
		et   *gcal.EventDateTime
		want string
	}{
		{"nil", nil, ""},
		{"all-day", &gcal.EventDateTime{Date: "2025-06-15"}, "2025-06-15"},
		{"datetime", &gcal.EventDateTime{DateTime: "2025-06-15T10:00:00-07:00"}, "2025-06-15T10:00:00-07:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEventTime(tt.et)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatAttendee(t *testing.T) {
	tests := []struct {
		name     string
		attendee *gcal.EventAttendee
		want     string
	}{
		{
			"basic",
			&gcal.EventAttendee{Email: "bob@example.com"},
			"bob@example.com",
		},
		{
			"with name and status",
			&gcal.EventAttendee{Email: "bob@example.com", DisplayName: "Bob", ResponseStatus: "accepted"},
			"Bob (bob@example.com) [accepted]",
		},
		{
			"organizer",
			&gcal.EventAttendee{Email: "alice@example.com", Organizer: true},
			"alice@example.com (organizer)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAttendee(tt.attendee)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseReminders(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := `[{"method":"popup","minutes":15},{"method":"email","minutes":30}]`
		reminders, err := parseReminders(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(reminders) != 2 {
			t.Fatalf("expected 2 reminders, got %d", len(reminders))
		}
		if reminders[0].Method != "popup" || reminders[0].Minutes != 15 {
			t.Errorf("first reminder: %+v", reminders[0])
		}
	})

	t.Run("empty", func(t *testing.T) {
		reminders, err := parseReminders("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reminders != nil {
			t.Errorf("expected nil, got %v", reminders)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		input := `[{"method":"sms","minutes":15}]`
		_, err := parseReminders(input)
		if err == nil {
			t.Error("expected error for invalid method")
		}
	})

	t.Run("invalid minutes", func(t *testing.T) {
		input := `[{"method":"popup","minutes":50000}]`
		_, err := parseReminders(input)
		if err == nil {
			t.Error("expected error for invalid minutes")
		}
	})
}

func TestBuildAttendees(t *testing.T) {
	emails := []string{"alice@example.com", "bob@example.com"}
	attendees := buildAttendees(emails)
	if len(attendees) != 2 {
		t.Fatalf("expected 2 attendees, got %d", len(attendees))
	}
	if attendees[0].Email != "alice@example.com" {
		t.Errorf("first attendee email = %q", attendees[0].Email)
	}
}

func TestBuildAttendeesEmpty(t *testing.T) {
	attendees := buildAttendees(nil)
	if attendees != nil {
		t.Errorf("expected nil for empty input, got %v", attendees)
	}
}

func TestEventToSummary(t *testing.T) {
	e := &gcal.Event{
		Id:       "evt123",
		Summary:  "Team Meeting",
		Location: "Room 101",
		Start:    &gcal.EventDateTime{DateTime: "2025-06-15T10:00:00Z"},
		End:      &gcal.EventDateTime{DateTime: "2025-06-15T11:00:00Z"},
		Status:   "confirmed",
		HtmlLink: "https://calendar.google.com/event?eid=evt123",
		Organizer: &gcal.EventOrganizer{
			Email:       "alice@example.com",
			DisplayName: "Alice",
		},
	}

	s := eventToSummary(e)
	if s.ID != "evt123" {
		t.Errorf("ID = %q, want %q", s.ID, "evt123")
	}
	if s.Summary != "Team Meeting" {
		t.Errorf("Summary = %q", s.Summary)
	}
	if s.Organizer != "Alice (alice@example.com)" {
		t.Errorf("Organizer = %q", s.Organizer)
	}
}
