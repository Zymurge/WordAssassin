package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	events "wordassassin/types/events"
)

func TestGetGame(t *testing.T) {
	// Setup
	target := GamePool{}
	require.NotNil(t, target)
	addGameToPool(t, &target, "g1", "some test code", "a file somewhere", "pass")
	addGameToPool(t, &target, "g2", "that same code", "a file somewhere else", "pass")
	actual, ok := target.GetGame("g2")
	require.True(t, ok, "An existing game should say it was fetched")
	require.NotNil(t, actual, "An existing game should have actually been fetched")
	require.Equal(t, "g2", actual.ID)
	require.Equal(t, "that same code", actual.GameCreator)
}

func TestGetGamesList(t *testing.T) {
	// Note: the test relies on sort orders by creation time for the final validation. Hence, the sleeps
	// to get past something where it wasn't always guaranteed to run in the add sequence ???
	target := GamePool{}
	require.NotNil(t, target)
	addGameToPool(t, &target, "g1", "alpha", "a file", "pass")
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, &target, "g2", "beta", "a file again", "pass")
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, &target, "g3", "gamma", "a file III", "pass")
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, &target, "g4", "a different greek letter", "a file strikes back", "pass")
	actual := target.GetGamesList()
	require.Equal(t, len(target.games), len(actual))
	require.Equal(t, "g3", actual[2].GetID())
}

func TestAddGame(t *testing.T) {
	target := GamePool{}
	t.Run("Positive", func(t *testing.T) {
		addGameToPool(t, &target, "add1", "test", "", "youshallnot")
	})
	t.Run("Duplicate ID", func(t *testing.T) {
		// add the same event twice to trigger dupe ID
		addGameToPool(t, &target, "add2", "testDupe", "", "youshallnot")
		addGameToPool(t, &target, "add2", "testDupe", "", "youshallnot", "duplicate")
	})
	t.Run("Missing ID", func(t *testing.T) {
		// create new event and break the ID field
		badEvent := NewGameFromEvent(events.GameCreatedEvent{
			ID:             "",
			TimeCreated:    time.Now(),
			EventType:      "GameCreatedEvent",
			GameCreator:    "creator",
			KillDictionary: "dict",
		})
		//badEvent.ID = ""
		err := target.AddGame(&badEvent)
		require.Errorf(t, err, "Missing ID should throw")
		require.Contains(t, err.Error(), "missing ID")
	})

}

func TestReconstitutePool(t *testing.T) {
	var eventsIn [4]*Game
	ev, _ := events.NewGameCreatedEvent("recon1", "testes", "killme", "Donner")
	g0 := NewGameFromEvent(ev)
	eventsIn[0] = &g0
	ev, _ = events.NewGameCreatedEvent("recon2", "testes", "killme", "Donner")
	g1 := NewGameFromEvent(ev)
	eventsIn[1] = &g1
	ev, _ = events.NewGameCreatedEvent("recon3", "testes", "killme", "Donner")
	g2 := NewGameFromEvent(ev)
	eventsIn[2] = &g2
	ev, _ = events.NewGameCreatedEvent("recon4", "testes", "killme", "Donner")
	g3 := NewGameFromEvent(ev)
	eventsIn[3] = &g3

	t.Run("Positive", func(t *testing.T) {
		target := &GamePool{}
		err := target.ReconstitutePool(eventsIn[:])
		require.NoError(t, err, "Expect success to be, well, successful")
		expected := "recon3"
		actual, ok := target.GetGame(expected)
		require.True(t, ok, "GetGame should return an OK")
		require.Equal(t, actual.GetID(), expected)
	})
	t.Run("DuplicateErrors", func(t *testing.T) {
		// change one member to create a dup
		// TODO: don't break the global array
		eventsIn[2].ID = "recon1"
		target := &GamePool{}
		err := target.ReconstitutePool(eventsIn[:])
		require.Error(t, err, "Should toss out an error for a duplicate")
		require.Contains(t, err.Error(), "duplicate", "Want to see that word in the error msg")
	})
}

//** Helper functions **//

// addGameToPool creates and adds a game to the pool. If an error is expected, it validates that it contains
// the optional passed in string. Otherwise, validates no error
func addGameToPool(t *testing.T, pool *GamePool, id, creator, dict, pass string, expectError ...string) {
	ev := NewGameFromEvent(events.GameCreatedEvent{
		ID:             id,
		TimeCreated:    time.Now(),
		EventType:      "GameCreatedEvent",
		GameCreator:    creator,
		KillDictionary: dict,
	})
	err := pool.AddGame(&ev)
	if len(expectError) > 0 {
		require.Error(t, err, "Wanted to see error adding to the test pool")
		require.Contains(t, err.Error(), expectError[0], "Wanted to see error adding to the test pool")
	} else {
		require.NoErrorf(t, err, "Didn't want to see error adding to the test pool: %v", err)
	}
}
