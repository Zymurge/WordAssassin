package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	events "wordassassin/types/events"
	"wordassassin/persistence"
)

func TestNewGamePool(t *testing.T) {
	// Create pool with some pre-existing games to be rehydrated during construction
	target, _ := getGamePoolWithMockMongo( t, 
		Game{
			ID:				"Mock1",
			TimeCreated:	time.Now(),
			GameCreator:	"Jim",
			KillDictionary:	"Websters",
			Status:			Starting,
			Passcode:		"Gandalf",
		},
		Game{
			ID:				"Mock2",
			TimeCreated:	time.Now(),
			GameCreator:	"Jimbo",
			KillDictionary:	"Websters",
			Status:			Starting,
			Passcode:		"Mitrandir",
		},	
	)
	require.NotNil(t, target)
	// validate that pre-existing games got hydrated
	actual, ok := target.GetGame("Mock2")
	require.True(t, ok, "An existing game should say it was fetched")
	require.NotNil(t, actual, "An existing game should have actually been fetched")
	require.Equal(t, "Mock2", actual.ID)
	require.Equal(t, "Jimbo", actual.GameCreator)
}

func TestNewGamePoolWithEmptyPreexistingList(t *testing.T) {
	// Create pool with some pre-existing games to be rehydrated during construction
	target, _ := getGamePoolWithMockMongo( t )
	require.NotNil(t, target)
}

func TestGetGame(t *testing.T) {
	// Setup
	target, _ := getGamePoolWithMockMongo(t)
	require.NotNil(t, target)
	addGameToPool(t, target, "g1", "some test code", "a file somewhere", "pass")
	addGameToPool(t, target, "g2", "that same code", "a file somewhere else", "pass")
	actual, ok := target.GetGame("g2")
	require.True(t, ok, "An existing game should say it was fetched")
	require.NotNil(t, actual, "An existing game should have actually been fetched")
	require.Equal(t, "g2", actual.ID)
	require.Equal(t, "that same code", actual.GameCreator)
}

func TestGetGamesList(t *testing.T) {
	// Note: the test relies on sort orders by creation time for the final validation. Hence, the sleeps
	// to get past something where it wasn't always guaranteed to run in the add sequence ???
	target, _ := getGamePoolWithMockMongo(t)
	require.NotNil(t, target)
	addGameToPool(t, target, "g1", "alpha", "a file", "pass")
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g2", "beta", "a file again", "pass")
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g3", "gamma", "a file III", "pass")
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g4", "a different greek letter", "a file strikes back", "pass")
	actual := target.GetGamesList()
	require.Equal(t, len(target.games), len(actual))
	require.Equal(t, "g3", actual[2].GetID())
}

func TestAddGame(t *testing.T) {
	target, _ := getGamePoolWithMockMongo(t)
	t.Run("Positive", func(t *testing.T) {
		addGameToPool(t, target, "add1", "test", "", "youshallnot")
	})
	t.Run("Duplicate ID", func(t *testing.T) {
		// add the same event twice to trigger dupe ID
		addGameToPool(t, target, "add2", "testDupe", "", "youshallnot")
		addGameToPool(t, target, "add2", "testDupe", "", "youshallnot", "duplicate")
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
		err := target.AddGame(badEvent)
		require.Errorf(t, err, "Missing ID should throw")
		require.Contains(t, err.Error(), "missing ID")
	})
}

func TestReconstitutePool(t *testing.T) {
	ev, _ := events.NewGameCreatedEvent("recon1", "testes", "killme", "Donner")
	g0 := NewGameFromEvent(ev)
	ev, _ = events.NewGameCreatedEvent("recon2", "testes", "killme", "Donner")
	g1 := NewGameFromEvent(ev)
	ev, _ = events.NewGameCreatedEvent("recon3", "testes", "killme", "Donner")
	g2 := NewGameFromEvent(ev)
	ev, _ = events.NewGameCreatedEvent("recon4", "testes", "killme", "Donner")
	g3 := NewGameFromEvent(ev)
	eventsIn := []Game{ g0, g1, g2, g3 }

	t.Run("Positive", func(t *testing.T) {
		target, _ := getGamePoolWithMockMongo(t)
		err := target.ReconstitutePool(eventsIn[:])
		require.NoError(t, err, "Expect success to be, well, successful")
		expected := "recon3"
		actual, ok := target.GetGame(expected)
		require.True(t, ok, "GetGame should return an OK")
		require.Equal(t, actual.GetID(), expected)
	})
	t.Run("DuplicateErrors", func(t *testing.T) {
		// create new pool with a dup
		dupEvents := []Game{ g1, g2, g1, g3, g0 }
		target, _ := getGamePoolWithMockMongo(t)
		err := target.ReconstitutePool(dupEvents[:])
		require.Error(t, err, "Should toss out an error for a duplicate")
		require.Contains(t, err.Error(), "duplicate", "Want to see that word in the error msg")
	})
}

//** Helper functions **//

// getGamePoolWithMockMongo creates a GamePool with a preset mock mongo and all positive mock behaviors
func getGamePoolWithMockMongo(t *testing.T, existingGames... persistence.Persistable) (target *GamePool, mm *persistence.MockMongoSession) {
	mm = &persistence.MockMongoSession{}
	mm.ConnectMode = "positive"
	mm.WriteMode = "positive"
	mm.QueryMode = "positive"
	// pre-existing games need to be added to the mock mongo before NewGamePool is called
	mm.FetchResults = existingGames
	target = NewGamePool(mm)
	return target, nil
}

// addGameToPool creates and adds a game to the pool. If an error is expected, it validates that it contains
// the optional passed in string. Otherwise, validates no error
func addGameToPool(t *testing.T, pool *GamePool, id, creator, dict, pass string, expectError ...string) {
	ev1,_ := events.NewGameCreatedEvent(id, creator, dict, pass)
	g1 := NewGameFromEvent(ev1)
	err := pool.AddGame(g1)
	if len(expectError) > 0 {
		require.Error(t, err, "Wanted to see error adding to the test pool")
		require.Contains(t, err.Error(), expectError[0], "Wanted to see error adding to the test pool")
	} else {
		require.NoErrorf(t, err, "Didn't want to see error adding to the test pool: %v", err)
	}
}
