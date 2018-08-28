package types

import (
	"testing"
	"github.com/stretchr/testify/require"
	"time"

	events "wordassassin/types/events"
)

func TestGetGame(t *testing.T) {
	// Setup
	target := GamePool{}
	require.NotNil(t, target)
	AddGameToPool(t, &target, "g1", "some test code", "a file somewhere", "pass" )
	AddGameToPool(t, &target, "g2", "that same code", "a file somewhere else", "pass" )
	actual, ok := target.GetGame("g2")
	require.True(t, ok, "An existing game should say it was fetched")
	require.NotNil(t, actual, "An existing game should have actually been fetched")
	require.Equal(t, "g2", actual.ID )
	require.Equal(t, "that same code", actual.GameCreator )
}

func TestGetGamesList(t *testing.T) {
	// Note: the test relies on sort orders by creation time for the final validation. Hence, the sleeps
	// to get past something where it wasn't always guaranteed to run in the add sequence ???
	target := GamePool{}
	require.NotNil(t, target)
	AddGameToPool(t, &target, "g1", "alpha", "a file", "pass" )
	time.Sleep(100 * time.Millisecond)
	AddGameToPool(t, &target, "g2", "beta", "a file again", "pass" )
	time.Sleep(100 * time.Millisecond)
	AddGameToPool(t, &target, "g3", "gamma", "a file III", "pass" )
	time.Sleep(100 * time.Millisecond)
	AddGameToPool(t, &target, "g4", "a different greek letter", "a file strikes back", "pass" )
	actual := target.GetGamesList()
	require.Equal(t, len(target.games), len(actual))
	require.Equal(t, "g3", actual[2].GetID())
}

func AddGameToPool(t *testing.T, pool *GamePool, id, creator, dict, pass string) {
	ev := NewGameFromEvent( events.GameCreatedEvent{
		ID:	         id,
		TimeCreated: time.Now(),
		EventType:	 "GameCreatedEvent",
		GameCreator: creator,
		KillDictionary: dict,
	} )
	err := pool.AddGame(&ev)
	require.NoErrorf(t, err, "Didn't want to see error adding to the test pool: %v", err)
}