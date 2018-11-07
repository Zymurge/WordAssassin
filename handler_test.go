package main

import (
	"bytes"
	"log"
//	"time"
	
	"testing"
	"github.com/stretchr/testify/require"
	
	"wordassassin/types"
	"wordassassin/types/events"
	dao "wordassassin/persistence"
)

// need to move this to a centralized place for tests
const (
	TestMongoURL   string = "localhost:27017"
	TestDbName     string = "testDB"
	TestCollection string = "TestCollection"
)

type mockControls struct {
	connectMode string       // positive
	writeMode   string       // positive
	returnVal   interface{}  // nil
}

type gameArgs struct {
	gameid		string // required
	creator 	string // required
	killdict 	string // default: afile.txt
	passcode 	string // default: melod
	numPlayers	int    // default: 7
	status  	types.GameStatus // default: Starting
}

type playerArgs struct {
	gameid 		string // required
	slackid 	string // required
	name		string // default: playerX
	email		string // default: pX@game.org
}	

// commandArgs used to send alternate args to commands than used to create games and players
// all are optional
type commandArgs struct {
	gameid      string
	creator     string
}

type testArgs struct {
	name    	string         // required
	h       	Handler        // the default testHandler
	wantErr 	bool           // required
	errText 	string         // ""
	gArgs    	gameArgs       // required for game tests
	pArgs    	playerArgs     // required for player tests
	cArgs       commandArgs    // default nil
	mock    	mockControls   // default mock controls
}


// TODO: add log validation to all tests
/*
func TestHandler_OnPlayerAdded(t *testing.T) {
	testHandler, mongo, blog := getHandlerWithMockMongoAndLogger()
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs {
		{"positive", testHandler, false, "", gameArgs{},
			playerArgs{"game1", "@fred", "fred", "fred@bedrock.org"},
			mockControls{"positive", "positive", nil}},
		{"PlayerPool error (from dup)", testHandler, true, "out of sync", gameArgs{},
			playerArgs{"game1", "@dupe_player", "dupey", "thesameguy@some.org"},
			mockControls{"positive", "positive", nil}},
		{"duplicate slackID (at mongo)", testHandler, true, "@fred already added", gameArgs{},
			playerArgs{"game1", "@fred", "fred", "fred@bedrock.org"},
			mockControls{"positive", "duplicate", nil}},
		{"empty slackID argument", testHandler, true, "missing SlackID", gameArgs{},
			playerArgs{"game1", "", "whatev", "bad@email.org"},
			mockControls{"positive", "positive", nil}},
		{"gameID doesn't exist", testHandler, true, "missing_gameid doesn't exist", gameArgs{},
			playerArgs{"missing_gameid", "@someone", "someone", "bad@email.org"},
			mockControls{"positive", "positive", nil}},
		{"gameID already started", testHandler, true, "not accepting players", gameArgs{},
			playerArgs{"started_gameid", "@toolate", "lagger", "tomorrow@procrastinate.org"},
			mockControls{"positive", "positive", nil}},
		{"empty GameID argument", testHandler, true, "doesn't exist", gameArgs{},
			playerArgs{"", "@someone", "someone", "bad@email.org"},
			mockControls{"positive", "positive", nil}},
		{"mongo fail", testHandler, true, "connect", gameArgs{},
			playerArgs{"game1", "n/a", "n/a", "bad@email.org"},
			mockControls{"no connect", "positive", nil}},
	}

	// Add a game named "game1" for cases that expect it
	game1 := types.Game{
		ID:             "game1",
		TimeCreated:    time.Now(),
		GameCreator:    "testmaster",
		KillDictionary: "websters",
		Status:         types.Starting,
	}
	testHandler.gPool.AddGame(game1)
	// Add a game named "started_gameid" for cases that expect it
	started := types.Game{
		ID:             "started_gameid",
		TimeCreated:    time.Now(),
		GameCreator:    "testmaster",
		KillDictionary: "websters",
		Status:         types.Playing,
	}
	testHandler.gPool.AddGame(started)
	// Add a preexisting player to support duplicate cases
	dupEv, _ := events.NewPlayerAddedEvent("game1", "@dupe_player", "", "")
	dupPlayer := types.NewPlayerFromEvent(dupEv)
	testHandler.pPool.AddPlayer(&dupPlayer)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongo.ConnectMode = tt.mock.connectMode
			mongo.WriteMode = tt.mock.writeMode
			err := tt.h.OnPlayerAdded(tt.pArgs.gameid, tt.pArgs.slackid, tt.pArgs.name, tt.pArgs.email)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual, err := testHandler.pPool.GetPlayer(tt.pArgs.gameid, tt.pArgs.slackid)
				require.NoErrorf(t, err, "Didn't want to see this: %v", err)
				require.Equal(t, tt.pArgs.name, actual.Name, "Didn't find player added despite success")
			}
		})
	}
}
*/
func TestHandler_OnGameCreated(t *testing.T) {
	// TODO: ensure tests for validating non-blank parameters
	testHandler, mongo, blog := getHandlerWithMockMongoAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs {
		testArgs {
			name: "positive",
			wantErr: false,
			gArgs: gameArgs {
				gameid: "game1",
				creator: "@fred",
				killdict: "notBlank",
				passcode: "notBlank",
				numPlayers: 2,
			},
			cArgs: commandArgs {
				gameid: "game1",
				creator: "@fred",
			},
		},
		testArgs {
			name: "empty gameID argument",
			wantErr: true,
			errText: "with blank gameid",
			gArgs: gameArgs {
				gameid: "",
				creator: "@someone",
				killdict: "notBlank",
				passcode: "notBlank",
				numPlayers: 3,
			},
			cArgs: commandArgs {
				gameid: "",
				creator: "",
			},
		},
		testArgs {
			name: "duplicate gameID (at mongo)",
			wantErr: true,
			errText: "dupe_game already created",
			gArgs: gameArgs {
				gameid: "dupe_game",
				creator: "@testmaster",
				killdict: "notBlank",
				passcode: "notBlank",
				numPlayers: 2,
			},
			mock: mockControls {
				connectMode: "positive",
				writeMode: "duplicate",
				returnVal: nil,
			},
			cArgs: commandArgs {
				gameid: "",
				creator: "",
			},
		},
		testArgs {
			name: "duplicate gameID (GamePool)",
			wantErr: true,
			errText: "GamePool out of sync",
			gArgs: gameArgs {
				gameid: "dupe_game",
				creator: "@testmaster",
				killdict: "notBlank",
				passcode: "notBlank",
				numPlayers: 2,
			},
			mock: mockControls {
				connectMode: "positive",
				writeMode: "positive",
				returnVal: nil,
			},
			cArgs: commandArgs {
				gameid: "",
				creator: "",
			},
		},
		testArgs {
			name: "mongo fail",
			wantErr: true,
			errText: "connect",
			gArgs: gameArgs {
				gameid: "fail",
				creator: "n/a",
				killdict: "notBlank",
				passcode: "notBlank",
				numPlayers: 0,
			},
			mock: mockControls {
				connectMode: "no connect",
				writeMode: "positive",
				returnVal: nil,
			},
			cArgs: commandArgs {
				gameid: "",
				creator: "",
			},
		},
	}

	// Add a game for tests that need a pre-existing id
	testHandler.gPool.AddGame( types.Game{
			ID:             "dupe_game",
			//TimeCreated:    time.Now(),
			GameCreator:    "@testmaster",
			KillDictionary: "websters",
			Status:         types.Starting,
		} )

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMockControlsFromArgs(mongo, tt.mock)
			err := testHandler.OnGameCreated(tt.gArgs.gameid, tt.gArgs.creator, tt.gArgs.killdict, tt.gArgs.passcode)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual, exists := testHandler.gPool.GetGame(tt.gArgs.gameid)
				require.NotNilf(t, exists, "Expected % to exist in the gamepool", tt.gArgs.gameid)
				require.Equal(t, tt.cArgs.gameid, actual.GetID(), "Didn't find game added despite success")
			}
		})
	}
}

func TestHandler_OnGameStarted(t *testing.T) {
	testHandler, mongo, blog := getHandlerWithMockMongoAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs {
		testArgs {
			name: "positive",
			wantErr: false,
			gArgs: gameArgs {
				gameid: "game1",
				creator: "@fred",
				numPlayers: 7,
			},
			cArgs: commandArgs {
				gameid: "game1",
				creator: "@fred",
			},
		},
		testArgs {
			name: "empty gameID argument",
			wantErr: true,
			errText: "requires a non-empty game ID and slack ID", 
			gArgs: gameArgs {
				gameid: "",
				creator: "@someone",
				numPlayers: 7,
			},
			cArgs: commandArgs {
				gameid: "",
				creator: "@someone",
			},
		},
		testArgs {
			name: "empty creator argument",
			wantErr: true,
			errText: "requires a non-empty game ID and slack ID", 
			gArgs: gameArgs {
				gameid: "game_filler",
				creator: "@someone",
				numPlayers: 7,
			},
			cArgs: commandArgs {
				gameid: "game_filler",
				creator: "",
			},
		},
		testArgs {
			name: "game state wrong",
			wantErr: true,
			errText: "state", 
			gArgs: gameArgs {
				gameid: "I'm running",
				creator: "@someone",
				status: types.Playing,
				numPlayers: 7,
			},
			cArgs: commandArgs{
				gameid: "I'm running",
				creator: "@someone",
			},
		},
		testArgs {
			name: "wrong creator",
			wantErr: true,
			errText: "creator", 
			gArgs: gameArgs{
				gameid: "hack_me",
				creator: "@da_man",
				status: types.Playing,
				numPlayers: 7,
			},
			cArgs: commandArgs {
				gameid: "hack_me",
				creator: "@a_hacker",
			},
		},
		testArgs {
			name: "mongo fail",
			wantErr: true,
			errText: "connect",
			gArgs: gameArgs {
				gameid: "mongo bad",
				creator: "@someone",
			},
			cArgs: commandArgs {
				gameid: "mongo bad",
				creator: "@someone",
			},
			mock: mockControls {
				connectMode: "no connect",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGame := newGameFromArgs(tt.gArgs)
			testHandler.gPool.AddGame(testGame)
			setMockControlsFromArgs(mongo, tt.mock)
			err := testHandler.OnGameStarted(tt.cArgs.gameid, tt.cArgs.creator)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual, exists := testHandler.gPool.GetGame(tt.gArgs.gameid)
				require.NotNilf(t, exists, "Expected % to exist in the gamepool", tt.gArgs.gameid)
				require.Equal(t, types.Playing, actual.Status, "Game status should have changed correctly")
			}
		})
	}
}


/*** Helpers ***/

func getHandlerWithMockMongoAndLogger(t *testing.T) (testHandler *Handler, mongo *dao.MockMongoSession, logBuf *bytes.Buffer) {
	mongo = dao.NewMockMongoSession()
	testPPool := types.PlayerPool{}
	testGPool := types.NewGamePool(mongo, &testPPool)
	logBuf = &bytes.Buffer{}
	logLabel := "handler_test: "
	blog := log.New(logBuf, logLabel, 0)
	testHandler = NewHandler(testGPool, &testPPool, mongo, blog)
	require.NotNil(t, testHandler, "Make sure creation worked")
	require.NotNil(t, testHandler.gPool, "Make sure we have a valid GamePool")
	require.NotNil(t, testHandler.pPool, "Make sure we have a valid PlayerPool")
	return
}

func newGameFromArgs(args gameArgs) types.Game {
/*
	gameid		string // required
	creator 	string // required
	killdict 	string // default: afile.txt
	passcode 	string // default: melod
	numPlayers	int    // default: 7
	status  	types.GameStatus // default: Starting
*/
	if args.killdict == "" { args.killdict = "afile.txt" }
	if args.passcode == "" { args.passcode = "melod" }
	if args.numPlayers == 0 { args.numPlayers = 7 }
	
	gce, _ := events.NewGameCreatedEvent(args.gameid, args.creator, args.killdict, args.passcode)
	myGame := types.NewGameFromEvent(gce)
	// TODO: validate that forcing the count is sufficient for tests or if mock playerpool is needed
	myGame.StartPlayers = args.numPlayers

	return myGame
}

func setMockControlsFromArgs(mongo *dao.MockMongoSession, args mockControls) {
	if args.connectMode == "" {
		mongo.ConnectMode = "positive"
	} else {
		mongo.ConnectMode = args.connectMode		
	}
	if args.writeMode == "" {
		mongo.WriteMode = "positive"
	} else {
		mongo.WriteMode = args.writeMode		
	}
}

/*
func createGameWithPlayers(h Handler, gameid string, numPlayers int) {
	myGame := types.NewGameFromEvent(
		events.NewGameCreatedEvent(gameid, "@testFunc", "some words", "sesame"),
	)
	h.gPool.AddGame(myGame)

	for i := 0; i < numPlayers; i++ {
		name := string("player#" + i)
		pae, _ := events.NewPlayerAddedEvent(gameid, "@" + name, name, name + "@some.org") 
		p := types.NewPlayerFromEvent(pae)
		h.pPool.AddPlayer(p)
	}
}
*/
