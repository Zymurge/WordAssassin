package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	events "wordassassin/types/events"
)

func TestNewGameFromEvent(t *testing.T) {
	ev, err := events.NewGameCreatedEvent( "bigape", "@King.Kong", "bananas.txt", "Jane")
	require.NoError(t, err, "Gotta get the game event built first")

	actual := NewGameFromEvent(ev)
	require.NotNil(t, actual, "Failed to instantiate")
	require.Equal(t, ev.GetID(), actual.ID)
	require.Equal(t, ev.GetTimeCreated(), actual.TimeCreated)
	require.Equal(t, ev.GameCreator, actual.GameCreator)
	require.Equal(t, ev.KillDictionary, actual.KillDictionary)
	require.Equal(t, ev.Passcode, actual.Passcode)
	require.Equal(t, Starting, actual.Status)
	require.Equal(t, time.Unix(0, 0), actual.StartTime)
	require.Equal(t, 0, actual.StartPlayers)
	require.Equal(t, 0, actual.RemainPlayers)
}

func TestDecode(t *testing.T) {
	require.FailNow(t, "Add me")
}

func TestGetID(t *testing.T) {
	require.FailNow(t, "Add me")
}

func TestGetStatus(t *testing.T) {
	require.FailNow(t, "Add me")
}

func TestGetStatusReport(t *testing.T) {
	require.FailNow(t, "Add me")
}

func TestStart(t *testing.T) {
	require.FailNow(t, "Add me")
}


	// ID             string     `json:"id" bson:"_id"`
	// TimeCreated    time.Time  `json:"timeCreated" bson:"timecreated"`
	// GameCreator    string     `json:"gameId" bson:"gameid"`
	// KillDictionary string     `json:"name" bson:"name"`
	// Passcode       string     `json:"passcode" bson:"passcode"`
	// Status         GameStatus `json:"status" bson:"status"`
	// StartTime      time.Time  `json:"starttime"`
	// StartPlayers   int        `json:"startplayers"`
	// RemainPlayers  int        `json:"remainplayers"`
