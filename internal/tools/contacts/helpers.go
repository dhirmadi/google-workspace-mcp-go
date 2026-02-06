package contacts

import (
	"fmt"
	"strings"

	"google.golang.org/api/people/v1"
)

// ContactSummary is a compact representation of a Google Contact.
type ContactSummary struct {
	ResourceName string   `json:"resource_name"`
	DisplayName  string   `json:"display_name,omitempty"`
	Emails       []string `json:"emails,omitempty"`
	Phones       []string `json:"phones,omitempty"`
	Organization string   `json:"organization,omitempty"`
	ETag         string   `json:"etag,omitempty"`
}

// ContactGroupSummary is a compact representation of a contact group.
type ContactGroupSummary struct {
	ResourceName string `json:"resource_name"`
	Name         string `json:"name"`
	MemberCount  int    `json:"member_count"`
	GroupType    string `json:"group_type"`
}

// personToSummary converts a People API Person to a contact summary.
func personToSummary(p *people.Person) ContactSummary {
	cs := ContactSummary{
		ResourceName: p.ResourceName,
		ETag:         p.Etag,
	}

	// Display name
	if len(p.Names) > 0 {
		cs.DisplayName = p.Names[0].DisplayName
	}

	// Emails
	for _, e := range p.EmailAddresses {
		cs.Emails = append(cs.Emails, e.Value)
	}

	// Phones
	for _, ph := range p.PhoneNumbers {
		cs.Phones = append(cs.Phones, ph.Value)
	}

	// Organization
	if len(p.Organizations) > 0 {
		org := p.Organizations[0]
		parts := make([]string, 0, 2)
		if org.Name != "" {
			parts = append(parts, org.Name)
		}
		if org.Title != "" {
			parts = append(parts, org.Title)
		}
		cs.Organization = strings.Join(parts, " — ")
	}

	return cs
}

// contactGroupToSummary converts a ContactGroup to a summary.
func contactGroupToSummary(g *people.ContactGroup) ContactGroupSummary {
	return ContactGroupSummary{
		ResourceName: g.ResourceName,
		Name:         g.Name,
		MemberCount:  int(g.MemberCount),
		GroupType:    g.GroupType,
	}
}

// formatContactDetail writes detailed contact info to a response builder.
func formatContactLine(cs ContactSummary) string {
	parts := []string{cs.DisplayName}
	if len(cs.Emails) > 0 {
		parts = append(parts, fmt.Sprintf("<%s>", cs.Emails[0]))
	}
	return strings.Join(parts, " ")
}

// personFieldsForRead returns the standard field mask for reading contacts.
func personFieldsForRead() string {
	return "names,emailAddresses,phoneNumbers,organizations,metadata"
}

// personFieldsForList returns the field mask for listing contacts.
func personFieldsForList() string {
	return "names,emailAddresses,phoneNumbers,organizations"
}

// buildPerson builds a Person object from contact creation/update inputs.
func buildPerson(givenName, familyName, email, phone, orgName, orgTitle string) *people.Person {
	person := &people.Person{}

	if givenName != "" || familyName != "" {
		person.Names = []*people.Name{
			{GivenName: givenName, FamilyName: familyName},
		}
	}

	if email != "" {
		person.EmailAddresses = []*people.EmailAddress{
			{Value: email},
		}
	}

	if phone != "" {
		person.PhoneNumbers = []*people.PhoneNumber{
			{Value: phone},
		}
	}

	if orgName != "" || orgTitle != "" {
		person.Organizations = []*people.Organization{
			{Name: orgName, Title: orgTitle},
		}
	}

	return person
}

func ptrBool(b bool) *bool { return &b }
