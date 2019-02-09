package types

import (
	"testing"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"

	events "wordassassin/types/events"
)

func TestGame(t *testing.T) {
	expectedID := "bigape"
	ev, err := events.NewGameCreatedEvent(expectedID, "@King.Kong", "bananas.txt", "Jane")
	mockPP := &MockPlayerPool{}
	require.NoError(t, err, "Gotta get the game event built first")
	actual := NewGameFromEvent(ev)
	t.Run("Validate New Game", func(t *testing.T) {
		require.NotNil(t, actual, "Failed to instantiate")
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, ev.GetTimeCreated(), actual.TimeCreated)
		require.Equal(t, ev.GameCreator, actual.GameCreator)
		require.Equal(t, ev.KillDictionary, actual.KillDictionary)
		require.Equal(t, ev.Passcode, actual.Passcode)
		require.Equal(t, Starting, actual.Status)
		require.Equal(t, time.Unix(0,0), actual.StartTime)
		require.Equal(t, 0, actual.StartPlayers)
		require.Equal(t, 0, actual.RemainPlayers)
	})
	t.Run("Decode", func(t *testing.T) {
		bytes, err := bson.Marshal(actual)
		require.NoError(t, err, "Failed to create bson")
		decoded := Game{}
		err = decoded.Decode(bytes)
		require.NoError(t, err, "Failed to Decode")
		require.Equal(t, actual.GetID(), decoded.ID)
		// TODO: file bug for truncated nsec on timestamp Unmarshal
		// require.Equal(t, actual.TimeCreated, decoded.TimeCreated)
	})
	t.Run("GetStatus", func(t *testing.T) {
		actualStatus := actual.GetStatus()
		require.Equal(t, Starting, actual.Status)
		require.Equal(t, "starting", actualStatus)
	})
	t.Run("GetStatusReport", func(t *testing.T) {
		actualStatus := actual.GetStatusReport()
		require.Contains(t, actualStatus, "Game Status for bigape")
		require.Contains(t, actualStatus, "Status: starting")
	})
	t.Run("GetAllPlayersInPool", func(t *testing.T) {
		mockPP.playersToReturn = generatePlayers(expectedID, 5)
		result := actual.GetPlayerList(mockPP)
		require.NotNil(t, result)
		require.Equal(t, 5, len(result))
	})
}

func TestStart(t *testing.T) {
	expectedID := "startableGame"
	ev, _ := events.NewGameCreatedEvent(expectedID, "@King.Kong", "bananas.txt", "Jane")
	youCanStartMeUp := NewGameFromEvent(ev)
	youCanStartMeUp.StartPlayers = 13
	players := generatePlayers(youCanStartMeUp.ID, 13)

	t.Run("Positive", func(t *testing.T) {
		err := youCanStartMeUp.Start(players)
		require.NoErrorf(t, err, "Didn't want to see this: %s", err)
		require.Equal(t, Playing, youCanStartMeUp.Status)
		actualStatus := youCanStartMeUp.GetStatus()
		require.Equal(t, "playing", actualStatus)
		// TODO: find a way to validate that timestamp was set to now
	})
	t.Run("Wrong state", func(t *testing.T) {
		runningGame := NewGameFromEvent(ev)
		runningGame.Status = Playing
		err := runningGame.Start(players)
		require.Error(t, err, "Expect an error when status is not 'starting'")
		require.Contains(t, err.Error(), "not in starting")
	})
	t.Run("Less than minimum players", func(t *testing.T) {
		minExpected := 8
		numPlayers := 5
		unfilledGame := NewGameFromEvent(ev)
		unfilledGame.MinimumPlayers = minExpected
		unfilledGame.StartPlayers = numPlayers
		err := unfilledGame.Start(players)
		require.Error(t, err, "Expect an error when status is not 'starting'")
		expectedMsg := fmt.Sprintf("Game requires %d players. Current count is %d", minExpected, numPlayers)
		require.Contains(t, err.Error(), expectedMsg)
	})
	t.Run("Invalid target list", func(t *testing.T) { // TODO: create a test when valid dictionary logic is defined
		require.True(t, true)
	})
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
func generatePlayers(gid string, numPlayers int) (players []*Player) {
	players = make([]*Player,numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = &Player{
			ID          :	gid + "@player#" + string(i),
			TimeCreated :	time.Now(),
			GameID		:	gid,
			Name        :	"I'm player # " + string(i),
			SlackID     :	"@slackid" + string(i),
			Email       :	"p" + string(i) + "@email.org",
			Status		:	Alive,
			Kills		:	0,
			Target		:	"",
			KillWord	:	"",
		}
	}
	return
}