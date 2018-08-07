package events

import (
	//"bytes"
	"testing"
	"time"

	bson "github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/require"
)

func TestNewGameCreatedEvent(t *testing.T) {
	expected := GameCreatedEvent{
		ID: 			"pastafarians unite",
		TimeCreated:	time.Date(2112, 11, 11, 10, 11, 12, 0, time.UTC),
		EventType:		"GameCreatedEvent",
		GameCreator:	"His Googly Appendages",
		KillDictionary:	"pirates of the ages",
		Passcode:		"fsm",
	}
	actual, err := NewGameCreatedEvent("pastafarians unite", "His Googly Appendages", "pirates of the ages", "fsm")
	require.NoError(t, err, "Positive ctor throws no errors")
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.EventType, actual.EventType)
	require.Equal(t, expected.GameCreator, actual.GameCreator)
	require.Equal(t, expected.KillDictionary, actual.KillDictionary)
	require.Equal(t, expected.Passcode, actual.Passcode)
}

func TestGameCreatedEvent_Decode(t *testing.T) {
		// Setup
		original := GameCreatedEvent{
			ID:				"testID",
			TimeCreated:	time.Date(2112, time.November, 13, 16, 20, 0, 0, time.UTC),
			EventType:		"GameCreatedEvent",
			GameCreator:	"@Bob_Marley",
			KillDictionary:	"websters",
		}
		asBytes, err := bson.Marshal(original)
		require.NoError(t, err, "Failure to marshal test object to bytes: %v", err)
		asRaw := bson.Raw{}
		err = bson.Unmarshal(asBytes,&asRaw)
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
	t.Run("Broken Mapping", func(t *testing.T) {
		// reset an expected string type to an int. Expect decode to err on this
		brokenRaw := asRaw
		brokenRaw.Data = append(brokenRaw.Data,[]byte(`junk`)[:]...)
		actual := &GameCreatedEvent{}
		err = actual.Decode(brokenRaw)
		require.Error(t, err, "Bad mapping should throw an error")
		require.Contains(t, err.Error(), "Cast issue", "Error message should indicate problem with casting")
		require.Contains(t, err.Error(), "KillDictionary", "Error message should indicate which element")
	}) /*
	t.Run("Missing tag", func(t *testing.T){
		// remove an expected k/v pair. Expect decode to err on this
		brokenMap := originalAsMap
		bytes.Trim(brokenMap.Data,"timecreated")
		//delete( brokenMap, "timecreated" )
		actual := &GameCreatedEvent{}
		err = actual.Decode(brokenMap)
		require.Error(t, err, "Missing key should throw an error")
		require.Contains(t, err.Error(), "Missing key", "Error message should indicate a key is missing")
		require.Contains(t, err.Error(), "timecreated", "Error message should indicate which element")
	}) */

}
