package events

import (
	"testing"
	"time"

	bson "github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/require"
)

func TestGameCreatedEvent_Decode(t *testing.T) {
		original := GameCreatedEvent{
			ID:				"testID",
			TimeCreated:	time.Date(2112, time.November, 13, 16, 20, 0, 0, time.UTC),
			GameCreator:	"@Bob_Marley",
			KillDictionary:	"websters",
		}
		asBSON, err := bson.Marshal(original)
		require.NoError(t, err, "Failure to marshal test object to BSON: %v", err)
		originalAsMap := bson.M{}
		err = bson.Unmarshal(asBSON, &originalAsMap)
		require.NoError(t, err, "Failure to unmarshal test object to map: %v", err)
	t.Run("Positive", func(t *testing.T){
		actual := &GameCreatedEvent{}
		err = actual.Decode(originalAsMap)
		require.NoError(t, err, "Failure to marshal test object to BSON: %v", err)
		require.Equal(t, original.ID, actual.ID)
		require.Equal(t, original.TimeCreated, actual.TimeCreated)
		require.Equal(t, original.GameCreator, actual.GameCreator)
		require.Equal(t, original.KillDictionary, actual.KillDictionary)
	})
	t.Run("Broken Mapping", func(t *testing.T){
		// reset an expected string type to an int. Expect decode to err on this
		brokenMap := originalAsMap
		brokenMap["killdictionary"] = 13
		actual := &GameCreatedEvent{}
		err = actual.Decode(brokenMap)
		require.Error(t, err, "Bad mapping should throw an error")
		require.Contains(t, err.Error(), "Cast issue", "Error message should indicate problem with casting")
		require.Contains(t, err.Error(), "KillDictionary", "Error message should indicate which element")
	})
	t.Run("Missing tag", func(t *testing.T){
		// remove an expected k/v pair. Expect decode to err on this
		brokenMap := originalAsMap
		delete( brokenMap, "timecreated" )
		actual := &GameCreatedEvent{}
		err = actual.Decode(brokenMap)
		require.Error(t, err, "Missing key should throw an error")
		require.Contains(t, err.Error(), "Missing key", "Error message should indicate a key is missing")
		require.Contains(t, err.Error(), "timecreated", "Error message should indicate which element")
	})

}

func TestPlayerAddedEvent_Decode(t *testing.T) {
		original := PlayerAddedEvent{
			ID:				"testID",
			TimeCreated:	time.Date(2112, time.February, 13, 16, 20, 0, 0, time.UTC),
			GameID:			"Redemption Song",
			Name:			"Bob Marley",
			SlackID:		"@wailers",
			Email:			"wailer@marley.com",
		}
		asBSON, err := bson.Marshal(original)
		require.NoError(t, err, "Failure to marshal test object to BSON: %v", err)
		asM := bson.M{}
		err = bson.Unmarshal(asBSON, &asM)
		require.NoError(t, err, "Failure to unmarshal test object to map: %v", err)
	t.Run("Positive", func(t *testing.T){
		actual := &PlayerAddedEvent{}
		err = actual.Decode(asM)
		require.NoError(t, err, "Failure to marshal test object to BSON: %v", err)
		require.Equal(t, original.ID, actual.ID)
		require.Equal(t, original.TimeCreated, actual.TimeCreated)
		require.Equal(t, original.Name, actual.Name)
		require.Equal(t, original.SlackID, actual.SlackID)
		require.Equal(t, original.Email, actual.Email)
	})

}