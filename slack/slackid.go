package slack

import (
	"fmt"
	"regexp"
)

// SlackID is used as a validated string that meets the Slack format
type SlackID string

// Validate checks that a string conforms to the Slack ID guidelines.
// It returns whether the string is valid and includes the reason if invalid.
func Validate(id string) (valid bool, reason error) {
	valid = true
	validSlackid := regexp.MustCompile(`^[U|W][a-zA-Z0=9]+`)
	if !validSlackid.MatchString(id) {
		valid = false
		reason = fmt.Errorf("A valid Slack ID must start with either a 'U' or 'W' and consist of only alphanumberics")
	}
	return
}

// New creates an instance of SlackID after validating that the string conforms to the Slack standards.
// If the string fails validation, the reason is return as an error and the SlackID returns an empty string.
func New(id string) (sID SlackID, reason error) {
	if valid, err := Validate(id); !valid {
		reason = err
	} else {
		sID = SlackID(id)
	}
	return
}

// NewInline is a test util that creates an instance of SlackID after validating that the string conforms to the Slack standards.
// It provides a single receiver form. If the string fails validation, the func panics.
func NewInline(id string) (SlackID) {
	sID, err := New(id)
	if(err != nil) {
		panic(fmt.Sprintf("Error creating SlackID for %s: %v", id, err))
	}
	return sID
}

// ToString provides the slack ID in string form
func (sID SlackID) ToString() string {
	return string(sID)
}
