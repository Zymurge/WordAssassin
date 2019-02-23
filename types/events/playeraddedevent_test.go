package events

import (
	"testing"
	"github.com/stretchr/testify/require"
	"bytes"
	"encoding/binary"
	"time"

	"wordassassin/persistence"
	"wordassassin/slack"
	
	"gopkg.in/mgo.v2/bson"
)

func TestPlayerAddedEventIsPersistable(t * testing.T) {
	ev := &PlayerAddedEvent {
		ID:				"I will persist",
		TimeCreated:	time.Date(2112, time.February, 13, 16, 20, 0, 0, time.UTC),
		EventType:		"PlayerAddedEvent",
		GameID:			"Time",
		Name:			"Pink Floyd",
	}
	
	_, ok := interface{}(ev).(persistence.Persistable)
	require.True(t, ok)
}

func TestNewPlayerAddedEvent_MultiplePermutations(t *testing.T) {
	tests := []struct {
		testname string
		ID string
		GameID string
		SlackID string
		Name string
		Email string
		wantErr bool
		msg string
	}{
		{
			"Positive_all",
			"game1+Uplayer1", "game1", "Uplayer1", "Joe", "joe@wa.com",
			false, "", 
		},
		{
			"Positive_blank_optionals",
			"game1+Uplayer1", "game1", "Uplayer1", "", "",
			false, "", 
		},
		{
			"No gameid",
			"missing GameID", "", "Uplayer1", "Joe", "joe@wa.com",
			true, "missing GameID field",
		},
		{
			"No slackid",
			"missing SlackID", "boo", "", "Joe", "joe@wa.com",
			true, "missing SlackID field",
		},
		{
			"Invalid slackid",
			"bad SlackID", "boo", "@UBADSLACK", "Joe", "joe@wa.com",
			true, "valid Slack ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			got, err := NewPlayerAddedEvent(tt.GameID, tt.SlackID, tt.Name, tt.Email)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.msg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.ID, got.GetID())
				require.Equal(t, tt.GameID, got.GameID)
				require.Equal(t, slack.SlackID(tt.SlackID), got.SlackID)
				require.Equal(t, "PlayerAddedEvent", got.EventType)
				require.NotNil(t, got.TimeCreated) // can't match a time now
				require.Equal(t, tt.Name, got.Name)
				require.Equal(t, tt.Email, got.Email)
			}
		})
	}
}

func TestNewPlayerAddedInline_Positive(t *testing.T) {
	expectedGameID := "inline_game"
	expectedSlackID := "UINLINETESTER"
	expectedID := expectedGameID + "+" + expectedSlackID
	require.NotPanics(t, func(){NewPlayerAddedInline(expectedGameID, expectedSlackID, "some dude", "email@addr.es")} )
	actual := NewPlayerAddedInline( expectedGameID, expectedSlackID, "some dude", "email@addr.es" )
	require.NotNil(t, actual, "Successful creation actually creates something")
	require.Equal(t, actual.GetID(), expectedID)
}

func TestNewPlayerAddedInline_Panics(t *testing.T) {
	require.Panics(t, func(){NewPlayerAddedInline("", "", "I panic", "email@addr.es")} )
}

func TestPlayerAddedEvent_GetTimeCreated(t *testing.T) {
	target := PlayerAddedEvent {
		ID:				"time check",
		TimeCreated:	time.Date(2112, time.February, 13, 16, 20, 0, 0, time.UTC),
		EventType:		"PlayerAddedEvent",
		GameID:			"Time",
		Name:			"Pink Floyd",
	}
	actual := target.GetTimeCreated()
	require.Equal(t, target.TimeCreated, actual)
}

func TestPlayerAddedEvent_Decode(t *testing.T) {
		original := PlayerAddedEvent {
			ID:				"testID",
			TimeCreated:	time.Date(2112, time.February, 13, 16, 20, 0, 0, time.Local),
			EventType:		"PlayerAddedEvent",
			GameID:			"Redemption Song",
			Name:			"Bob Marley",
			SlackID:		slack.SlackID("WAILERS"),
			Email:			"wailer@marley.com",
		}
		asBytes, err := bson.Marshal(original)
		require.NoError(t, err, "Failure to marshal test object to bytes: %v", err)
	
	t.Run("Positive", func(t *testing.T) {
		actual := &PlayerAddedEvent{}
		err = actual.Decode(asBytes)
		require.NoError(t, err, "Failure to Decode BSON: %v", err)
		require.Equal(t, original.ID, actual.ID)
		require.Equal(t, original.TimeCreated, actual.TimeCreated)
		require.Equal(t, original.Name, actual.Name)
		require.Equal(t, original.SlackID, actual.SlackID)
		require.Equal(t, original.Email, actual.Email)
	} )
	// reset an expected string type to an int. Expect decode to err on this
	// doing this via manipulating byte arrays is a bitch!
	t.Run("Broken Mapping", func(t *testing.T) {
		badValue := new(bytes.Buffer)
		err := binary.Write(badValue, binary.LittleEndian, int32(13))
		require.NoError(t, err, "Failure to create byte array for bad value %v", err)
		// leverage the clean byte array from setup to make a copy with the bad value
		brokenBytes := bytes.Replace(asBytes, []byte(`PlayerAddedEvent`), badValue.Bytes(), 1)

		actual := &PlayerAddedEvent{}
		err = actual.Decode(brokenBytes)
		require.Error(t, err, "Bad mapping should throw an error")
		require.Contains(t, err.Error(), "corrupted", "Error message should indicate problem with casting")
	})
}