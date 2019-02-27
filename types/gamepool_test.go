package types

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	events "wordassassin/types/events"
	"wordassassin/persistence"
	"wordassassin/slack"
)

func TestNewGamePool(t *testing.T) {
	// Create pool with some pre-existing games to be rehydrated during construction
	target, _ := getGamePoolWithMockMongo( t, nil,
		&Game{
			ID:				"Mock1",
			TimeCreated:	time.Now(),
			GameCreator:	slack.SlackID("WJim"),
			KillDictionary:	"Websters",
			Status:			Starting,
			Passcode:		"Gandalf",
		},
		&Game{
			ID:				"Mock2",
			TimeCreated:	time.Now(),
			GameCreator:	slack.SlackID("UJimbo"),
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
	require.Equal(t, "UJimbo", actual.GameCreator.ToString())
}

func TestNewGamePoolWithEmptyPreexistingList(t *testing.T) {
	// Create pool with some pre-existing games to be rehydrated during construction
	target, _ := getGamePoolWithMockMongo( t, nil )
	require.NotNil(t, target)
}

func TestGetGameFunc(t *testing.T) {
	// Setup
	target, _ := getGamePoolWithMockMongo(t, nil)
	require.NotNil(t, target)
	// Note: the test relies on sort orders by creation time for the final validation. Hence, the sleeps
	// to get past something where it wasn't always guaranteed to run in the add sequence ???
	addGameToPool(t, target, "g1", "Ualpha", "a file", "pass", 1)
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g2", "Ubeta", "a file again", "pass", 1)
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g3", "Ugamma", "a file III", "pass", 1)
	time.Sleep(100 * time.Millisecond)
	addGameToPool(t, target, "g4", "UaDifferentGreekLetter", "a file strikes back", "pass", 1)
	t.Run("GetGame: positive", func(t *testing.T) {
		actual, ok := target.GetGame("g2")
		require.True(t, ok, "An existing game should say it was fetched")
		require.NotNil(t, actual, "An existing game should have actually been fetched")
		require.Equal(t, "g2", actual.ID)
		require.Equal(t, "Ubeta", actual.GameCreator.ToString())
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
	target, _ := getGamePoolWithMockMongo(t, nil)
	t.Run("Positive", func(t *testing.T) {
		addGameToPool(t, target, "add1", "Utest", "", "youshallnot", 0)
	})
	t.Run("Duplicate ID", func(t *testing.T) {
		// add the same event twice to trigger dupe ID
		addGameToPool(t, target, "add2", "UtestDupe", "", "youshallnot", 0)
		addGameToPool(t, target, "add2", "UtestDupe", "", "youshallnot", 0, "duplicate")
	})
	t.Run("Missing ID", func(t *testing.T) {
		// create new event and break the ID field
		badEvent := NewGameFromEvent(events.GameCreatedEvent{
			ID:             "",
			TimeCreated:    time.Now(),
			EventType:      "GameCreatedEvent",
			GameCreator:    "Ucreator",
			KillDictionary: "dict",
		})
		err := target.AddGame(&badEvent)
		require.Errorf(t, err, "Missing ID should throw")
		require.Contains(t, err.Error(), "missing ID")
	})
}

func TestAddPlayerToGame(t *testing.T) {
	myGameID := "playeradderer"
	mockPP := &MockPlayerPool{}
	target, _ := getGamePoolWithMockMongo(t, mockPP)
	require.NotNil(t, target)
	gm := addGameToPool(t, target, myGameID, "Uplayervacuum", "a file", "pass", 1)

	t.Run("Positive", func(t *testing.T) {
		err := target.AddPlayerToGame(myGameID, events.PlayerAddedEvent{ ID: "yo"})
		require.NoError(t, err, "Positive tests throw no errors")
		require.Equal(t, gm.StartPlayers, 2, "StartingPlayer count should increment on player add")
	})
	t.Run("Missing game", func(t *testing.T) {
		err := target.AddPlayerToGame("Who, me?", events.PlayerAddedEvent{})
		require.Error(t, err, "Should get an error on the game id check failure")
		require.Contains(t, err.Error(), "GameID: Who, me? doesn't exist", "Tell us why it broke")
	})
	t.Run("Duplicate player", func(t *testing.T){
		mockPP.AddPlayerError = "mock error: duplicate ID"
		err := target.AddPlayerToGame(myGameID, events.PlayerAddedEvent{ ID: "whatev"})
		require.Error(t, err, "Should get an error on the game id check failure")
		expectedErr := fmt.Sprintf("PlayerPool: attempt to add duplicate player: %s in game: %s", "whatev", myGameID)
		require.Contains(t, err.Error(), expectedErr, "Tell us why it broke")
	})
	t.Run("PlayerPool issue (not duplicate)", func(t *testing.T){
		mockPP.AddPlayerError = "mock error: bad bad stuff happened"
		err := target.AddPlayerToGame(myGameID, events.PlayerAddedEvent{ ID: "whatev"})
		require.Error(t, err, "Should get an error on the game id check failure")
		require.Contains(t, err.Error(), "PlayerPool: ", "Tell us where it broke")
		require.Contains(t, err.Error(), mockPP.AddPlayerError, "Tell us what broke")
	})
}

func TestCanAddPlayer(t *testing.T) {
	target, _ := getGamePoolWithMockMongo(t, nil)
	require.NotNil(t, target)
	addGameToPool(t, target, "good", "Ualpha", "a file", "pass", 1)
	gm := addGameToPool(t, target, "playingGame", "Ubeta", "a file again", "pass", 8)
	gm.Status = Playing

	t.Run("Positive", func(t *testing.T) {
		result, err := target.CanAddPlayers("good")
		require.True(t, result, "Game should be accepting players")
		require.NoError(t, err, "Error should not be set for true return value")
	})
	t.Run("Game not found", func(t *testing.T) {
		result, err := target.CanAddPlayers("missingGame")
		require.False(t, result, "Missing game should not be accepting players")
		require.Error(t, err, "Error should be set for false return value")
		require.Contains(t, err.Error(), "missingGame doesn't exist", "Error message should mention missing gameid")
	})
	t.Run("Game state not Starting", func(t *testing.T) {
		result, err := target.CanAddPlayers("playingGame")
		require.False(t, result, "Game not in Starting state should not be accepting players")
		require.Error(t, err, "Error should be set for false return value")
		require.Contains(t, err.Error(), "playingGame is not accepting players. State=playing", "Error message should mention incorrect state")
	})
}

func TestReconstitutePool(t *testing.T) {
	g0 := NewGameFromEvent(events.NewGameCreatedInline("recon1", "Utestes", "killme", "Donner"))
	g1 := NewGameFromEvent(events.NewGameCreatedInline("recon2", "Utestes", "killme", "Donner"))
	g2 := NewGameFromEvent(events.NewGameCreatedInline("recon3", "Utestes", "killme", "Donner"))
	g3 := NewGameFromEvent(events.NewGameCreatedInline("recon4", "Utestes", "killme", "Donner"))
	eventsIn := []*Game{ &g0, &g1, &g2, &g3 }

	t.Run("Positive", func(t *testing.T) {
		target, _ := getGamePoolWithMockMongo(t, nil)
		err := target.ReconstitutePool(eventsIn[:])
		require.NoError(t, err, "Expect success to be, well, successful")
		expected := "recon3"
		actual, ok := target.GetGame(expected)
		require.True(t, ok, "GetGame should return an OK")
		require.Equal(t, actual.GetID(), expected)
	})
	t.Run("DuplicateErrors", func(t *testing.T) {
		// create new pool with a dup
		dupEvents := []*Game{ &g1, &g2, &g1, &g3, &g0 }
		target, _ := getGamePoolWithMockMongo(t, nil)
		err := target.ReconstitutePool(dupEvents[:])
		require.Error(t, err, "Should toss out an error for a duplicate")
		require.Contains(t, err.Error(), "duplicate", "Want to see that word in the error msg")
	})
}

func TestStartGame(t *testing.T) {
	// Setup: create a game, some players, a playerpool (mock) and finally the gamepool
	myGameID := "add1"
	myCreator, sErr := slack.New("UdaStarter")
	require.NoError(t, sErr, "Blew up in creating slack id")
	myGame := &Game{
		ID:				myGameID,
		TimeCreated:	time.Now(),
		GameCreator:	myCreator,
		KillDictionary:	"wordz",
		Status:			Starting,
		Passcode:		"MickJ",
		// FYI: a valid game currently requires at least 5 players to start
		StartPlayers:	6,
	}
	players := makePlayerList(t, myGameID, 6)
	mockPP := &MockPlayerPool{ playersToReturn: players }
	target, _ := getGamePoolWithMockMongo(t, mockPP, myGame)

	t.Run("Positive", func(t *testing.T) {
		// Need to grab the reconsituted instance after restore from mock mongo
		targetGame, ok := target.GetGame(myGameID)
		require.True(t, ok, "Couldn't find reconstituted game instance for ID: %s", myGameID)
		err := target.StartGame(myGameID, myCreator)
		require.NoError(t, err)
		require.Equal(t, Playing, targetGame.Status, "Once started, the game should have the correct status")
		// return status to reuse
		targetGame.Status = Starting
	})
	t.Run("Blank slackid", func(t *testing.T) {
		err := target.StartGame("", myCreator)
		require.Error(t, err, "Should get an error on a blank slack id")
		require.Contains(t, err.Error(), "requires a non-empty game ID and creator ID", "Tell us why it broke")
	})
	t.Run("Blank creator", func(t *testing.T) {
		err := target.StartGame("whatev", "")
		require.Error(t, err, "Should get an error on a blank creator")
		require.Contains(t, err.Error(), "requires a non-empty game ID and creator ID", "Tell us why it broke")
	})
	t.Run("Missing game", func(t *testing.T) {
		err := target.StartGame("Who, me?", myCreator)
		require.Error(t, err, "Should get an error on the game id check failure")
		require.Contains(t, err.Error(), "GameID: Who, me? doesn't exist", "Tell us why it broke")
	})
	t.Run("Wrong game state", func(t *testing.T) {
		myStartedGame := addGameToPool(t, target, "startedGame", "UGAMEBREAKER", "wordz", "MickJ", 6)
		myStartedGame.Status = Playing
		err := target.StartGame(myStartedGame.ID, myStartedGame.GameCreator)
		require.Error(t, err, "Should get an error on starting a game not in the Starting state")
		require.Contains(t, err.Error(), "GameID: startedGame is not accepting players", "Tell us why it broke")
	})
	t.Run("Wrong creator", func(t *testing.T) {
		someoneElsesGame := addGameToPool(t, target, "lockDown", "UGAMEBREAKER", "wordz", "MickJ", 6)
		err := target.StartGame(someoneElsesGame.ID, slack.NewInline("UNOTME"))
		require.Error(t, err, "Should get an error on starting a game with the wrong creator ID")
		require.Contains(t, err.Error(), "GameID: lockDown cannot be started by non-creator", "Tell us why it broke")
	})
	t.Run("PlayerPool issue", func(t *testing.T){
		mockPP.GetPlayerError = "mock error: bad bad stuff happened"
		err := target.StartGame(myGameID, myCreator)
		require.Error(t, err, "Should get an error when PlayerPool gets the player list")
		require.Contains(t, err.Error(), "PlayerPool: ", "Tell us where it broke")
		require.Contains(t, err.Error(), mockPP.GetPlayerError, "Tell us what broke")
		// restore mock
		mockPP.GetPlayerError = ""
	})

}

//** Helper functions **//

// addGameToPool creates and adds a game to the GamePool. If an error is expected, it validates that it contains
// the optional passed in string. Otherwise, validates no error
func addGameToPool(t *testing.T, pool *GamePool, id, creator, dict, pass string, numPlayers int, expectError ...string) *Game {
	g1 := NewGameFromEvent(events.NewGameCreatedInline(id, creator, dict, pass))
	g1.StartPlayers = numPlayers
	err := pool.AddGame(&g1)
	if len(expectError) > 0 {
		require.Error(t, err, "Wanted to see error adding to the test pool")
		require.Contains(t, err.Error(), expectError[0], "Wanted to see error adding to the test pool")
	} else {
		require.NoErrorf(t, err, "Didn't want to see error adding to the test pool: %v", err)
	}
	return &g1
}

// getGamePoolWithMockMongo creates a GamePool with a preset mock mongo set for all positive mock behaviors
// uses either the passed in PlayerPoolAbstraction, or if nil, creates a default instance of PlayerPool
func getGamePoolWithMockMongo(t *testing.T, pp PlayerPoolAbstraction, existingGames... persistence.Persistable) (target *GamePool, mm *persistence.MockMongoSession) {
	mm = &persistence.MockMongoSession{}
	if pp == nil {
		pp = &PlayerPool{}
	}
	mm.ConnectMode = "positive"
	mm.WriteMode = "positive"
	mm.QueryMode = "positive"
	// pre-existing games need to be added to the mock mongo before NewGamePool is called, since it relies
	// on reconstitution from the mongo layer for existing games
	// ** warning: games will be new instances
	mm.FetchResults = existingGames
	target = NewGamePool(mm, pp)
	return target, nil
}

func makePlayerList(t * testing.T, gameid string, numPlayers int) []*Player {
	players := make([]*Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		id    := fmt.Sprintf("Uname%d", i)
		name  := fmt.Sprintf("name%d", i)
		email := fmt.Sprintf("iam%d@mail.org", i)
		if pl, err := NewPlayer(gameid, id, name, email); err != nil {
			require.NoError(t, err, "Issue creating player for PlayerList")
		} else {
			players[i] = &pl
		}
	}
	require.Equal(t, numPlayers, len(players), "Somehow built the wrong number of players")
	return players
}
