package events

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewPlayerAddedEvent(t *testing.T) {
	expected := PlayerAddedEvent{
		ID:          "a game+@slackdude",
		TimeCreated: time.Date(2112, 11, 11, 10, 11, 12, 0, time.UTC),
		GameID:      "a game",
		SlackID:     "@slackdude",
		Name:        "a name",
		Email:       "my@email.org",
	}
	actual, err := NewPlayerAddedEvent("a game", "@slackdude", "a name", "my@email.org")
	require.NoError(t, err, "Positive ctor throws no errors")
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.GameID, actual.GameID)
	require.Equal(t, expected.SlackID, actual.SlackID)
	require.Equal(t, expected.Name, actual.Name)
}

func TestNewPlayerAddedEventMultiple(t *testing.T) {
	tests := []struct {
		testname string
		gameid string
		slackid string
		name string
		email string
		wantErr bool
		msg string
		id string
	}{
		{
			"Positive_all",
			"game1", "@player1", "Joe", "joe@wa.com",
			false, "", "game1+@player1",
		},
		{
			"Positive_blank_optionals",
			"game1", "@player1", "", "",
			false, "", "game1+@player1",
		},
		{
			"No gameid",
			"", "@player1", "Joe", "joe@wa.com",
			true, "missing GameID", "",
		},
		{
			"No slackid",
			"boo", "", "Joe", "joe@wa.com",
			true, "missing SlackID", "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPlayerAddedEvent(tt.gameid, tt.slackid, tt.name, tt.email)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.msg)
			} else {
				require.NoError(t, err)
				require.Equal(t, got.GameID, tt.gameid)
				require.Equal(t, got.SlackID, tt.slackid)
				require.Equal(t, got.ID, tt.id)
				require.NotNil(t, got.TimeCreated) // can't match a time now
			}
		})
	}

}
func TestPlayerAddedEvent_GetTimeCreated(t *testing.T) {
	tests := []struct {
		name string
		e    *PlayerAddedEvent
		want time.Time
	}{
		{
			"test1",
			&PlayerAddedEvent{
				TimeCreated: time.Unix(13, 0),
			},
			time.Unix(13, 0),
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetTimeCreated(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlayerAddedEvent.GetTimeCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}
