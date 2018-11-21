package types

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	events "wordassassin/types/events"
	"wordassassin/persistence"
)

func TestNewGamePool(t *testing.T) {
	// Create pool with some pre-existing games to be rehydrated during construction
	target, _ := getGamePoolWithMockMongo( t, 
		&Game{
			ID:				"Mock1",
			TimeCreated:	time.Now(),
			GameCreator:	"Jim",
			KillDictionary:	"Websters",
			Status:			Starting,
			Passcode:		"Gandalf",
		},
		&Game{
			ID:				"Mock2",
			TimeCreated:	time.Now(),
			GameCreator:	"Jimbo",
			KillDictionary:	"Websters",
			Status:			Starting,
			Passcode:		"Mithrandir",
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

func TestAllGetGameFunc(t *testing.T) {
	// Setup
	target, _ := getGamePoolWithMockMongo(t)
	require.NotNil(t, target)
	// Note: the test relies on sort orders by creation time for the final validation. Hence, the sleeps
	// to get past something where it wasn't always guaranteed to run in the add sequence ???
	addGameToPool(t, target, "g1", "alpha", "a file", "pass", 1)
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g2", "beta", "a file again", "pass", 1)
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g3", "gamma", "a file III", "pass", 1)
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g4", "a different greek letter", "a file strikes back", "pass", 1)
	t.Run("GetGame: positive", func(t *testing.T) {
		actual, ok := target.GetGame("g2")
		require.True(t, ok, "An existing game should say it was fetched")
		require.NotNil(t, actual, "An existing game should have actually been fetched")
		require.Equal(t, "g2", actual.ID)
		require.Equal(t, "beta", actual.GameCreator)
	})
	t.Run("GetGame: not found", func(t *testing.T) {
		actual, ok := target.GetGame("say what?")
		require.False(t, ok, "Don't lie about what can't be found")
		require.Nil(t, actual, "Don't make it easy to make mistakes")
	})
	t.Run("GetGamesList: positive", func(t *testing.T) {
		actual := target.GetGamesList()
		require.Equal(t, len(target.games), len(actual))
		require.Equal(t, "g3", actual[2].GetID())
	})

}

func TestAddGame(t *testing.T) {
	target, _ := getGamePoolWithMockMongo(t)
	t.Run("Positive", func(t *testing.T) {
		addGameToPool(t, target, "add1", "test", "", "youshallnot", 0)
	})
	t.Run("Duplicate ID", func(t *testing.T) {
		// add the same event twice to trigger dupe ID
		addGameToPool(t, target, "add2", "testDupe", "", "youshallnot", 0)
		addGameToPool(t, target, "add2", "testDupe", "", "youshallnot", 0, "duplicate")
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
//	dummyPP := &PlayerPool{}
	g0 := NewGameFromEvent(events.NewGameCreatedInline("recon1", "testes", "killme", "Donner"))
	g1 := NewGameFromEvent(events.NewGameCreatedInline("recon2", "testes", "killme", "Donner"))
	g2 := NewGameFromEvent(events.NewGameCreatedInline("recon3", "testes", "killme", "Donner"))
	g3 := NewGameFromEvent(events.NewGameCreatedInline("recon4", "testes", "killme", "Donner"))
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

func TestStartGame(t *testing.T) {
	// Setup: create gamepool, playerpool, a game and some players for the game in the pool
	myGameID := "add1"
	myCreator := "@daStarter"
	// Create a pool of players with incremental attributes
	// FYI: a valid game currently requires 5 players to start
	players := makePlayerList(t, myGameID, 6)
	mockPP := MockPlayerPool{ playersToReturn: players }
	// Create the mockmongo session ... 
	// TODO: enhance the helper method to optionally take a PlayerPool mock, and deprecate this multiline setup
	mm := &persistence.MockMongoSession{}
	mm.ConnectMode = "positive"
	mm.WriteMode = "positive"
	mm.QueryMode = "positive"
	// mm.FetchResults = existingGames
	// Create the GamePool using the mocks
	target := NewGamePool(mm, mockPP)
	addGameToPool(t, target, myGameID, myCreator, "", "youshallnot", 6)
	actualGame, ok := target.GetGame(myGameID)
	require.True(t, ok, "Error getting back the mapped game")

	t.Run("Positive", func(t *testing.T){
		err := target.StartGame(myGameID, myCreator)
		require.NoError(t, err)
		require.Equal(t, Playing, actualGame.Status, "Once started, the game should have the correct status")
	})
	// game doesn't exists
	// game in correct state for start
	// game creator matches
	// game start returns an error

}

//** Helper functions **//

// getGamePoolWithMockMongo creates a GamePool with a preset mock mongo and all positive mock behaviors
func getGamePoolWithMockMongo(t *testing.T, existingGames... persistence.Persistable) (target *GamePool, mm *persistence.MockMongoSession) {
	mm = &persistence.MockMongoSession{}
	dummyPP := &PlayerPool{}
	mm.ConnectMode = "positive"
	mm.WriteMode = "positive"
	mm.QueryMode = "positive"
	// pre-existing games need to be added to the mock mongo before NewGamePool is called
	mm.FetchResults = existingGames
	target = NewGamePool(mm, dummyPP)
	return target, nil
}

// addGameToPool creates and adds a game to the GamePool. If an error is expected, it validates that it contains
// the optional passed in string. Otherwise, validates no error
func addGameToPool(t *testing.T, pool *GamePool, id, creator, dict, pass string, numPlayers int, expectError ...string) {
	g1 := NewGameFromEvent(events.NewGameCreatedInline(id, creator, dict, pass))
	g1.StartPlayers = numPlayers
	err := pool.AddGame(g1)
	if len(expectError) > 0 {
		require.Error(t, err, "Wanted to see error adding to the test pool")
		require.Contains(t, err.Error(), expectError[0], "Wanted to see error adding to the test pool")
	} else {
		require.NoErrorf(t, err, "Didn't want to see error adding to the test pool: %v", err)
	}
}

func makePlayerList(t * testing.T, gameid string, numPlayers int) []*Player {
	players := make([]*Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		id    := fmt.Sprintf("@name%d", i)
		name  := fmt.Sprintf("name%d", i)
		email := fmt.Sprintf("iam%d@mail.org", i)
		pl := NewPlayerFromEvent(events.NewPlayerAddedInline(gameid, id, name, email))
		players[i] = &pl
	}
	require.Equal(t, numPlayers, len(players), "Somehow built the wrong number of players")
	return players
}
