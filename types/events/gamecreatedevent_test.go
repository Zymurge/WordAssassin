package events

import (
	"testing"
	"github.com/stretchr/testify/require"
	"bytes"
	"encoding/binary"
	"time"
	bson "github.com/globalsign/mgo/bson"
)

func TestNewGameCreatedEventMultiple(t *testing.T) {
	tests := []struct {
		testname string
		ID string
		GameCreator string
		KillDictionary string
		Passcode string
		wantErr bool
		msg string
	}{
		{
			"Positive_all",
			"The Game", "@queen", "lyrics", "dragonattack",
			false, "", 
		},
		{
			"Positive_blank_optionals",
			"The Game", "@queen", "lyrics", "",
			false, "", 
		},
		{
			"No gameid",
			"", "@queen", "lyrics", "dragonattack",
			true, "",
		},
		{
			"No creator",
			"Not Quite The Game", "", "lyrics", "dragonattack",
			true, "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			got, err := NewGameCreatedEvent(tt.ID, tt.GameCreator, tt.KillDictionary, tt.Passcode)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.msg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.ID, got.ID)
				require.Equal(t, "GameCreatedEvent", got.EventType)
				require.NotNil(t, got.TimeCreated) // can't match a time now
				require.Equal(t, tt.GameCreator, got.GameCreator)
				require.Equal(t, tt.KillDictionary, got.KillDictionary)
				require.Equal(t, tt.Passcode, got.Passcode)
			}
		})
	}
}

func TestGameCreatedEvent_GetTimeCreated(t *testing.T) {
	target := GameCreatedEvent {
		ID:				"time check",
		TimeCreated:	time.Date(2112, time.February, 13, 16, 20, 0, 0, time.UTC),
		EventType:      "GameCreatedEvent",
		GameCreator:    "@Bob_Marley",
		KillDictionary: "websters",
	}
	actual := target.GetTimeCreated()
	require.Equal(t, target.TimeCreated, actual)
}

func TestGameCreatedEvent_Decode(t *testing.T) {
	// Setup
	original := GameCreatedEvent{
		ID:             "testID",
		TimeCreated:    time.Date(2112, time.November, 13, 16, 20, 0, 0, time.UTC),
		EventType:      "GameCreatedEvent",
		GameCreator:    "@Bob_Marley",
		KillDictionary: "websters",
	}
	asBytes, err := bson.Marshal(original)
	require.NoError(t, err, "Failure to marshal test object to bytes: %v", err)
	asRaw := bson.Raw{}
	err = bson.Unmarshal(asBytes, &asRaw)
	require.NoError(t, err, "Failure to unmarshal test object to bson.Raw: %v", err)

	t.Run("Positive", func(t *testing.T) {
		actual := &GameCreatedEvent{}
		err = actual.Decode(asRaw)
		require.NoError(t, err, "Failure to marshal test object to BSON: %v", err)
		require.Equal(t, original.ID, actual.ID)
		require.Equal(t, original.TimeCreated, actual.TimeCreated)
		require.Equal(t, original.EventType, actual.EventType)
		require.Equal(t, original.GameCreator, actual.GameCreator)
		require.Equal(t, original.KillDictionary, actual.KillDictionary)
		require.Equal(t, original.Passcode, actual.Passcode)
	})
	// reset an expected string type to an int. Expect decode to err on this
	// doing this via manipulating byte arrays is a bitch!
	t.Run("Broken Mapping", func(t *testing.T) {
		badValue := new(bytes.Buffer)
		err := binary.Write(badValue, binary.LittleEndian, int32(13))
		require.NoError(t, err, "Failure to create byte array for bad value %v", err)
		// leverage the clean byte array from setup to make a copy with the bad value and then create a new bson.Raw from it
		brokenBytes := bytes.Replace(asBytes, []byte(`GameCreatedEvent`), badValue.Bytes(), 1)
		brokenRaw := bson.Raw{}
		err = bson.Unmarshal(brokenBytes, &brokenRaw)
		require.NoError(t, err, "Failure to unmarshal test object to bson.Raw: %v", err)

		actual := &GameCreatedEvent{}
		err = actual.Decode(brokenRaw)
		require.Error(t, err, "Bad mapping should throw an error")
		// error message from unmarshall is "Document is corrupted" -- passing that through is good enough for now
		require.Contains(t, err.Error(), "corrupted", "Error message should indicate problem with casting")
	})
}