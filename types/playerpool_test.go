package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	events "wordassassin/types/events"
)

func TestPlayerPool(t *testing.T) {
	// Setup
	target := PlayerPool{}
	require.NotNil(t, target)
	p1 := addPlayerToPool(t, &target, "game1", "@Joe", "Joe", "joe@wa.org")
	p2 := addPlayerToPool(t, &target, "game2", "@Joe", "Joe", "joe@wa.org")
	addPlayerToPool(t, &target, "game3", "@Joe", "Joe", "joe@wa.org")
	addPlayerToPool(t, &target, "game1", "@Jim", "Jim", "jim@wa.org")
	addPlayerToPool(t, &target, "game1", "@Josh", "Josh", "josh@wa.org")

	// Execute
	t.Run("AddPlayer: Duplicate ID", func(t *testing.T) {
		err := target.AddPlayer(&p1)
		require.Error(t, err, "Should have thrown on duplicate")
		require.Contains(t, err.Error(), "duplicate", "Error must mention the issue")
	})
	t.Run("AddPlayer: Missing ID", func(t *testing.T) {
		p3 := p1
		p3.ID = ""
		err := target.AddPlayer(&p3)
		require.Error(t, err, "Should have thrown on missing ID")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetPlayerByID: positive", func(t *testing.T) {
		actual, err := target.GetPlayerByID(p2.GetID())
		require.NoErrorf(t, err, "Positive tosses no errors, but this one did!: %v", err)
		require.Equal(t, &p2, actual)
	})
	t.Run("GetPlayerByID: not found", func(t *testing.T) {
		_, err := target.GetPlayerByID("bad ID")
		require.Errorf(t, err, "Error you will If ID you miss")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetPlayerByID: before players initialized", func(t *testing.T) {
		badTarget := PlayerPool{}
		_, err := badTarget.GetPlayerByID("who cares?")
		require.Errorf(t, err, "Not sure what happens here")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetPlayer: positive", func(t *testing.T) {
		actual, err := target.GetPlayer(p2.GameID, p2.SlackID)
		require.NoErrorf(t, err, "Positive tosses no errors, but this one did!: %v", err)
		require.Equal(t, &p2, actual)
	})
	t.Run("GetPlayer: not found", func(t *testing.T) {
		_, err := target.GetPlayer("No me existe", p2.SlackID)
		require.Errorf(t, err, "Error you will If ID you miss")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetPlayer: blank ID", func(t *testing.T) {
		_, err := target.GetPlayer("", p2.SlackID)
		require.Errorf(t, err, "Error you will If ID you miss")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetAllPlayersInGame: positive", func(t *testing.T) {
		actual, err := target.GetAllPlayersInGame(p1.GameID)
		require.NoErrorf(t, err, "Positive tosses no errors, but this one did!: %v", err)
		require.Equal(t, 3, len(actual), "GetAll count should be all of them")
	})

}

// *** Helpers *** //

// addPlayerToPool creates and adds a player to the PlayerPool. If an error is expected, it validates that it contains
// the optional passed in string. Otherwise, validates no error
func addPlayerToPool(t *testing.T, pool *PlayerPool, gameid, slackid, name, email string, expectError ...string) Player {
	ev := events.NewPlayerAddedInline(gameid, slackid, name, email)
	//require.NoErrorf(t, err, "Error on addPlayerToPool creation: %v", err)
	player := NewPlayerFromEvent(ev)
	err := pool.AddPlayer(&player)
	if len(expectError) > 0 {
		require.Error(t, err, "Wanted to see error adding to the test pool")
		require.Contains(t, err.Error(), expectError[0], "Wanted to see error adding to the test pool")
	} else {
		require.NoErrorf(t, err, "Didn't want to see error adding to the test pool: %v", err)
	}
	return player
}
