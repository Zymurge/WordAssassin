package main

import (
	"bytes"
	"log"

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

type mongoControls struct {
	connectMode  string       // positive
	writeMode    string       // positive
	returnVal    interface{}  // nil
}

type gPoolControls struct {
	gamesList		*[]types.Game
	returnVal		interface{}  // nil
	addGameErr 		string       // default: ""
	startGameErr	string       // default: ""
}

type gameArgs struct {
	gameid		 string // required
	creator 	 string // required
	killdict 	 string // default: afile.txt
	passcode 	 string // default: melod
	numPlayers	 int    // default: 7
	status  	 types.GameStatus // default: Starting
}

type playerArgs struct {
	gameid 		 string // required
	slackid 	 string // required
	name		 string // default: playerX
	email		 string // default: pX@game.org
}	

// commandArgs used to send alternate args to commands than used to create games and players
// all are optional
type commandArgs struct {
	gameid       string
	creator      string
}

type testArgs struct {
	name    	string         // required
	h       	Handler        // the default testHandler
	wantErr 	bool           // required
	errText 	string         // ""
	gArgs    	gameArgs       // required for game tests
	pArgs    	playerArgs     // required for player tests
	cArgs       commandArgs    // default nil
	mongoCtrl  	mongoControls  // default mock controls
	gPoolCtrl	gPoolControls  // default gPool controls
}

func TestHandlerCtorPositive(t *testing.T) {
	mongo := dao.NewMockMongoSession()
	testPPool := types.PlayerPool{}
	testGPool := types.NewGamePool(mongo, &testPPool)
	logBuf := &bytes.Buffer{}
	logLabel := "handler_ctortest: "
	blog := log.New(logBuf, logLabel, 0)
	result := NewHandler(testGPool, &testPPool, mongo, blog)
	require.NotNil(t, result, "Make sure creation worked")
	require.NotNil(t, result.gPool, "Make sure we have a valid GamePool")
	require.NotNil(t, result.pPool, "Make sure we have a valid PlayerPool")
	require.Contains(t, logBuf.String(), logLabel)
	require.Contains(t, logBuf.String(), "Handler created")
}

func TestHandlerCtorNilPointers(t *testing.T) {
	mongo := dao.NewMockMongoSession()
	testPPool := types.PlayerPool{}
	testGPool := types.NewGamePool(mongo, &testPPool)
	logBuf := &bytes.Buffer{}
	logLabel := "handler_ctortest: "
	blog := log.New(logBuf, logLabel, 0)

	require.PanicsWithValue(t, "GamePool argument is nil", func(){ NewHandler(nil, &testPPool, mongo, blog) })
	require.PanicsWithValue(t, "PlayerPool argument is nil", func(){ NewHandler(testGPool, nil, mongo, blog) })
	require.PanicsWithValue(t, "MongoSession argument is nil", func(){ NewHandler(testGPool, &testPPool, nil, blog) })
	require.PanicsWithValue(t, "Logger argument is nil", func(){ NewHandler(testGPool, &testPPool, mongo, nil) })
}


// TODO: add log validation to all tests
/*
func TestHandler_OnPlayerAdded(t *testing.T) {
	testHandler, mongo, blog := getHandlerWithMockMongoAndLogger()
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs {
		{"positive", testHandler, false, "", gameArgs{},
			playerArgs{"game1", "@fred", "fred", "fred@bedrock.org"},
			mongoControls{"positive", "positive", nil}},
		{"PlayerPool error (from dup)", testHandler, true, "out of sync", gameArgs{},
			playerArgs{"game1", "@dupe_player", "dupey", "thesameguy@some.org"},
			mongoControls{"positive", "positive", nil}},
		{"duplicate slackID (at mongo)", testHandler, true, "@fred already added", gameArgs{},
			playerArgs{"game1", "@fred", "fred", "fred@bedrock.org"},
			mongoControls{"positive", "duplicate", nil}},
		{"empty slackID argument", testHandler, true, "missing SlackID", gameArgs{},
			playerArgs{"game1", "", "whatev", "bad@email.org"},
			mongoControls{"positive", "positive", nil}},
		{"gameID doesn't exist", testHandler, true, "missing_gameid doesn't exist", gameArgs{},
			playerArgs{"missing_gameid", "@someone", "someone", "bad@email.org"},
			mongoControls{"positive", "positive", nil}},
		{"gameID already started", testHandler, true, "not accepting players", gameArgs{},
			playerArgs{"started_gameid", "@toolate", "lagger", "tomorrow@procrastinate.org"},
			mongoControls{"positive", "positive", nil}},
		{"empty GameID argument", testHandler, true, "doesn't exist", gameArgs{},
			playerArgs{"", "@someone", "someone", "bad@email.org"},
			mongoControls{"positive", "positive", nil}},
		{"mongo fail", testHandler, true, "connect", gameArgs{},
			playerArgs{"game1", "n/a", "n/a", "bad@email.org"},
			mongoControls{"no connect", "positive", nil}},
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
			mongo.ConnectMode = tt.mongoCtrl.connectMode
			mongo.WriteMode = tt.mongoCtrl.writeMode
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

func TestHandler_OnPlayerAdded(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs {
		testArgs { name: "positive",
			wantErr: false,
			gArgs: gameArgs {
				gameid: "game1",
				creator: "@fred",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			cArgs: commandArgs {
				gameid: "game1",
				creator: "@fred",
			},
		},
	}
	// Add player(s) for tests that require pre-existing players
	testHandler.pPool.AddPlayer( types.NewPlayer( events.NewPlayerAddedEvent(
		"targetGame", "@firstdupe", "George", "me@you.net") ) )

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMongoControlsFromArgs(mongo, tt.mongoCtrl)
			setGPoolControlsFromArgs(gPool, tt.gPoolCtrl)
			err := testHandler.OnGameCreated(tt.gArgs.gameid, tt.gArgs.creator, tt.gArgs.killdict, tt.gArgs.passcode)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual, exists := testHandler.gPool.GetGame(tt.gArgs.gameid)
				require.NotNilf(t, exists, "Expected ID: %s to exist in the gamepool", tt.gArgs.gameid)
				require.Equal(t, tt.cArgs.gameid, actual.GetID(), "Didn't find game added despite success")
			}
		})
	}

}

func TestHandler_OnGameCreated(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs {
		testArgs { name: "positive",
			wantErr: false,
			gArgs: gameArgs {
				gameid: "game1",
				creator: "@fred",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			cArgs: commandArgs {
				gameid: "game1",
				creator: "@fred",
			},
		},
		testArgs { name: "empty gameID argument",
			wantErr: true,
			errText: "with blank gameid",
			gArgs: gameArgs {
				gameid: "",
				creator: "@notblank",
				killdict: "notBlank",
				passcode: "notBlank",
			},
		},
		testArgs { name: "empty creator argument",
			wantErr: true,
			errText: "with blank creator",
			gArgs: gameArgs {
				gameid: "notblank",
				creator: "",
				killdict: "notBlank",
				passcode: "notBlank",
			},
		},
		testArgs { name: "empty killdict argument",
			wantErr: true,
			errText: "with blank killdict",
			gArgs: gameArgs {
				gameid: "notblank",
				creator: "@notblank",
				killdict: "",
				passcode: "notBlank",
			},
		},
		testArgs { name: "empty passcode argument",
			wantErr: true,
			errText: "with blank passcode",
			gArgs: gameArgs {
				gameid: "notblank",
				creator: "@notblank",
				killdict: "notBlank",
				passcode: "",
			},
		},
		testArgs { name: "force NewGameCreatedEvent error",
			wantErr: true,
			errText: "must start with '@'",
			gArgs: gameArgs {
				gameid: "notblank",
				creator: "invalid: missing @",
				killdict: "notBlank",
				passcode: "notBlank",
			},
		},
		testArgs { name: "duplicate gameID (at mongo)",
			wantErr: true,
			errText: "dupe_game already created",
			gArgs: gameArgs {
				gameid: "dupe_game",
				creator: "@testmaster",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls {
				connectMode: "positive",
				writeMode: "duplicate",
				returnVal: nil,
			},
		},
		testArgs { name: "duplicate gameID (GamePool)",
			wantErr: true,
			errText: "GamePool out of sync",
			gArgs: gameArgs {
				gameid: "dupe_game",
				creator: "@testmaster",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls {
				connectMode: "positive",
				writeMode: "positive",
				returnVal: nil,
			},
			gPoolCtrl: gPoolControls {
				addGameErr: "duplicate",
			},
		},
		testArgs { name: "mongo fail",
			wantErr: true,
			errText: "connect",
			gArgs: gameArgs {
				gameid: "fail",
				creator: "@Ifailoften",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls {
				connectMode: "no connect",
				writeMode: "positive",
				returnVal: nil,
			},
		},
		testArgs { name: "GamePool unknown error",
			wantErr: true,
			errText: "Issue on GameCreated add to GamePool: mock error",
			gArgs: gameArgs {
				gameid: "Bob",
				creator: "@Blahblah",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls {
				connectMode: "positive",
				writeMode: "positive",
				returnVal: nil,
			},
			gPoolCtrl: gPoolControls {
				addGameErr: "mock error",
			},
		},
	}

	// Add a game for tests that need a pre-existing id
	gPool.AddGame( &types.Game{
			ID:             "dupe_game",
			//TimeCreated:    time.Now(),
			GameCreator:    "@testmaster",
			KillDictionary: "websters",
			Status:         types.Starting,
		} )

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMongoControlsFromArgs(mongo, tt.mongoCtrl)
			setGPoolControlsFromArgs(gPool, tt.gPoolCtrl)
			err := testHandler.OnGameCreated(tt.gArgs.gameid, tt.gArgs.creator, tt.gArgs.killdict, tt.gArgs.passcode)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual, exists := testHandler.gPool.GetGame(tt.gArgs.gameid)
				require.NotNilf(t, exists, "Expected ID: %s to exist in the gamepool", tt.gArgs.gameid)
				require.Equal(t, tt.cArgs.gameid, actual.GetID(), "Didn't find game added despite success")
			}
		})
	}
}

func TestHandler_OnGameStarted(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
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
			name: "GamePool returns an error",
			wantErr: true,
			errText: "mock GamePool error message", 
			gArgs: gameArgs {
				gameid: "error me",
				creator: "@someone",
			},
			cArgs: commandArgs {
				gameid: "error me",
				creator: "@someone",
			},
			mongoCtrl: mongoControls {
				connectMode: "no connect",
			},
			gPoolCtrl: gPoolControls {
				startGameErr: "mock GamePool error message",
			},
		},
		// TODO: move the following test cases to GamePool.StartGame()
/*		
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
			mongoCtrl: mongoControls {
				connectMode: "no connect",
			},
		},
	*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGame := newGameFromArgs(tt.gArgs)
			gPool.AddGame(&testGame)
			setMongoControlsFromArgs(mongo, tt.mongoCtrl)
			setGPoolControlsFromArgs(gPool, tt.gPoolCtrl)
			err := testHandler.OnGameStarted(tt.cArgs.gameid, tt.cArgs.creator)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
			}
		})
	}
}


/*** Helpers ***/

func getHandlerWithMocksAndLogger(t *testing.T) (testHandler *Handler, mockMongo *dao.MockMongoSession, mockGPool *types.MockGamePool, logBuf *bytes.Buffer) {
	mockMongo = dao.NewMockMongoSession()
	testPPool := types.PlayerPool{}
	mockGames := []*types.Game {
		&types.Game{ ID: "mockity", GameCreator: "God" },
	}
	mockGPool = &types.MockGamePool{
		GamesToReturn: mockGames,
	}
	logBuf = &bytes.Buffer{}
	logLabel := "handler_test: "
	blog := log.New(logBuf, logLabel, 0)
	testHandler = NewHandler(mockGPool, &testPPool, mockMongo, blog)
	require.NotNil(t, testHandler, "Make sure creation worked")
	require.NotNil(t, testHandler.gPool, "Make sure we have a valid GamePool")
	require.NotNil(t, testHandler.pPool, "Make sure we have a valid PlayerPool")
	return
}

// newGameFromArgs builds a Game for test purposes using a mix of required and optional fields with default values.
// gameid		string // required
// creator 		string // required
// killdict 	string // default: afile.txt
// passcode 	string // default: melod
// numPlayers	int    // default: 7
// status  		types.GameStatus // default: Starting
func newGameFromArgs(args gameArgs) types.Game {
	if args.killdict == "" { args.killdict = "afile.txt" }
	if args.passcode == "" { args.passcode = "melod" }
	if args.numPlayers == 0 { args.numPlayers = 7 }
	
	gce, _ := events.NewGameCreatedEvent(args.gameid, args.creator, args.killdict, args.passcode)
	myGame := types.NewGameFromEvent(gce)
	// TODO: validate that forcing the count is sufficient for tests or if mock playerpool is needed
	myGame.StartPlayers = args.numPlayers

	return myGame
}

func setGPoolControlsFromArgs(gpool *types.MockGamePool, args gPoolControls) {
	gpool.AddGameError = args.addGameErr
	gpool.StartGameError = args.startGameErr
}


func setMongoControlsFromArgs(mongo *dao.MockMongoSession, args mongoControls) {
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
