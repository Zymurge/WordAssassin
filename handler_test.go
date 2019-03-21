package main

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	dao "wordassassin/persistence"
	"wordassassin/slack"
	"wordassassin/types"
	"wordassassin/types/events"
)

// need to move this to a centralized place for tests
const (
	TestMongoURL   string = "localhost:27017"
	TestDbName     string = "testDB"
	TestCollection string = "TestCollection"
)

type mongoControls struct {
	connectMode string      	// positive
	writeMode   string      	// positive
	returnVal   interface{} 	// nil
}

type gPoolControls struct {
	gamesList    []*types.Game
	returnVal    interface{} 	// nil
	addGameErr   string      	// default: ""
	addPlayerErr string      	// default: ""
	canAddErr    string      	// default: ""
	getGameErr   string      	// default: ""
	startGameErr string      	// default: ""
}

type gameArgs struct {
	gameid     string           // required
	creator    string           // required
	killdict   string           // default: afile.txt
	passcode   string           // default: melod
	numPlayers int              // default: 7
	status     types.GameStatus // default: Starting
}

type playerArgs struct {
	gameid  string 				// required
	slackid string 				// required
	name    string 				// default: playerX
	email   string 				// default: pX@game.org
}

// commandArgs used to send alternate args to commands than used to create games and players
// all are optional
type commandArgs struct {
	gameid  string
	creator string
}

type testArgs struct {
	name      string        	// required
	h         Handler       	// the default testHandler
	wantErr   bool          	// required
	errText   string        	// ""
	gArgs     gameArgs      	// required for game tests
	pArgs     playerArgs    	// required for player tests
	cArgs     commandArgs   	// default nil
	mongoCtrl mongoControls 	// default mock controls
	gPoolCtrl gPoolControls 	// default gPool controls
}

func TestHandlerCtorPositive(t *testing.T) {
	mongo := dao.NewMockMongoSession()
	testPPool := types.PlayerPool{}
	testGPool := types.NewGamePool(mongo, &testPPool)
	logBuf := &bytes.Buffer{}
	logLabel := "handler_ctortest: "
	blog := log.New(logBuf, logLabel, 0)
	result := NewHandler(testGPool, mongo, blog)
	require.NotNil(t, result, "Make sure creation worked")
	require.NotNil(t, result.gPool, "Make sure we have a valid GamePool")
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

	require.PanicsWithValue(t, "GamePool argument is nil", func() { NewHandler(nil, mongo, blog) })
	require.PanicsWithValue(t, "MongoSession argument is nil", func() { NewHandler(testGPool, nil, blog) })
	require.PanicsWithValue(t, "Logger argument is nil", func() { NewHandler(testGPool, mongo, nil) })
}

func TestHandler_OnPlayerAdded(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs{
		testArgs{name: "positive",
			wantErr: false,
			pArgs: playerArgs{
				gameid:  "game1",
				slackid: "UFRED",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
		},
		testArgs{name: "invalid game ID",
			wantErr: true,
			errText: "OnPlayerAdded: The request is missing GameID",
			pArgs: playerArgs{
				gameid:  "",
				slackid: "UCREATEDME",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
		},
		testArgs{name: "invalid slack ID",
			wantErr: true,
			errText: "OnPlayerAdded: Player does not have a valid Slack ID",
			pArgs: playerArgs{
				gameid:  "game1",
				slackid: "notValid",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
		},
		testArgs{name: "invalid game state",
			wantErr: true,
			errText: "OnPlayerAdded: game startedgame: (mock) message about state",
			pArgs: playerArgs{
				gameid:  "startedgame",
				slackid: "UCREATEDME",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
			gPoolCtrl: gPoolControls{
				canAddErr: "(mock) message about state",
			},
		},
		testArgs{name: "duplicate players",
			wantErr: true,
			errText: "OnPlayerAdded: Player UALREADYEXIST already added to game Dupity",
			pArgs: playerArgs{
				gameid:  "Dupity",
				slackid: "UALREADYEXIST",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
			mongoCtrl: mongoControls{
				connectMode: "positive",
				writeMode:   "duplicate",
				returnVal:   nil,
			},
		},
		testArgs{name: "mongo issue",
			wantErr: true,
			errText: "OnPlayerAdded: Mongodb write issue: Mock error on write",
			pArgs: playerArgs{
				gameid:  "Dupity",
				slackid: "UALREADYEXIST",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
			mongoCtrl: mongoControls{
				connectMode: "positive",
				writeMode:   "fail",
				returnVal:   nil,
			},
		},
		testArgs{name: "gamepool issue",
			wantErr: true,
			errText: "OnPlayerAdded: (mock) gamepool puke",
			pArgs: playerArgs{
				gameid:  "Dupity",
				slackid: "UALREADYEXIST",
				name:    "fred",
				email:   "fred@bedrock.org",
			},
			gPoolCtrl: gPoolControls{
				addPlayerErr: "(mock) gamepool puke",
			},
		},
	}
	// Add player(s) for tests that require pre-existing players
	preexist, err := events.NewPlayerAddedEvent("targetGame", "UFIRSTDUPE", "George", "me@you.net")
	require.NoError(t, err, "Failure to create preexisting player in setup")
	testHandler.gPool.AddPlayerToGame("targetGame", preexist)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMongoControlsFromArgs(mongo, tt.mongoCtrl)
			setGPoolControlsFromArgs(gPool, tt.gPoolCtrl)
			err := testHandler.OnPlayerAdded(tt.pArgs.gameid, tt.pArgs.slackid, tt.pArgs.name, tt.pArgs.email)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), "OnPlayerAdded:", "All errors should start with the func name", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				require.Equal(t, tt.pArgs.gameid, gPool.PlayerAdded.GameID, "OnPlayerAdded mock did not recieve the correct gameid" )
				require.Equal(t, tt.pArgs.slackid, gPool.PlayerAdded.Event.SlackID.ToString(), "OnPlayerAdded mock did not recieve the correct slackid in the event")
			}
		})
	}
}

func TestHandler_OnGameCreated(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs{
		testArgs{name: "positive",
			wantErr: false,
			gArgs: gameArgs{
				gameid:   "game1",
				creator:  "UFRED",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			cArgs: commandArgs{
				gameid:  "game1",
				creator: "UFRED",
			},
		},
		testArgs{name: "empty gameID argument",
			wantErr: true,
			errText: "OnGameCreated: The request is missing GameID",
			gArgs: gameArgs{
				gameid:   "",
				creator:  "UNOTBLANK",
				killdict: "notBlank",
				passcode: "notBlank",
			},
		},
		testArgs{name: "empty creator argument",
			wantErr: true,
			errText: "OnGameCreated: A valid Slack ID",
			gArgs: gameArgs{
				gameid:   "notblank",
				creator:  "",
				killdict: "notBlank",
				passcode: "notBlank",
			},
		},
		testArgs{name: "empty killdict argument",
			wantErr: true,
			errText: "OnGameCreated: The request is missing Killdict",
			gArgs: gameArgs{
				gameid:   "notblank",
				creator:  "UNOTBLANK",
				killdict: "",
				passcode: "notBlank",
			},
		},
		testArgs{name: "empty passcode argument",
			wantErr: true,
			errText: "OnGameCreated: The request is missing Passcode",
			gArgs: gameArgs{
				gameid:   "notblank",
				creator:  "UNOTBLANK",
				killdict: "notBlank",
				passcode: "",
			},
		},
		testArgs{name: "force NewGameCreatedEvent error",
			wantErr: true,
			errText: "valid Slack ID",
			gArgs: gameArgs{
				gameid:   "notblank",
				creator:  "invalid format",
				killdict: "notBlank",
				passcode: "notBlank",
			},
		},
		testArgs{name: "duplicate gameID (at mongo)",
			wantErr: true,
			errText: "dupe_game already created",
			gArgs: gameArgs{
				gameid:   "dupe_game",
				creator:  "UTESTMASTER",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls{
				connectMode: "positive",
				writeMode:   "duplicate",
				returnVal:   nil,
			},
		},
		testArgs{name: "duplicate gameID (GamePool)",
			wantErr: true,
			errText: "GamePool out of sync",
			gArgs: gameArgs{
				gameid:   "dupe_game",
				creator:  "UTESTMASTER",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls{
				connectMode: "positive",
				writeMode:   "positive",
				returnVal:   nil,
			},
			gPoolCtrl: gPoolControls{
				addGameErr: "duplicate",
			},
		},
		testArgs{name: "mongo fail",
			wantErr: true,
			errText: "connect",
			gArgs: gameArgs{
				gameid:   "fail",
				creator:  "UWILLFAIL",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls{
				connectMode: "no connect",
				writeMode:   "positive",
				returnVal:   nil,
			},
		},
		testArgs{name: "GamePool unknown error",
			wantErr: true,
			errText: "Issue on GameCreated add to GamePool: mock error",
			gArgs: gameArgs{
				gameid:   "Bob",
				creator:  "UWILLFAIL",
				killdict: "notBlank",
				passcode: "notBlank",
			},
			mongoCtrl: mongoControls{
				connectMode: "positive",
				writeMode:   "positive",
				returnVal:   nil,
			},
			gPoolCtrl: gPoolControls{
				addGameErr: "mock error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMongoControlsFromArgs(mongo, tt.mongoCtrl)
			setGPoolControlsFromArgs(gPool, tt.gPoolCtrl)
			err := testHandler.OnGameCreated(tt.gArgs.gameid, tt.gArgs.creator, tt.gArgs.killdict, tt.gArgs.passcode)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), "OnGameCreated:", "All errors should start with the func name", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				require.Equal(t, tt.gArgs.gameid, gPool.GameAdded.Added.ID, "AddGame mock did not recieve the correct gameid" )
				require.Equal(t, tt.gArgs.creator, gPool.GameAdded.Added.GameCreator.ToString(), "AddGame mock did not recieve the correct creator" )
			}
		})
	}
}

func TestHandler_OnGameStarted(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	tests := []testArgs{
		testArgs{name: "positive",
			wantErr: false,
			gArgs: gameArgs{
				gameid:     "game1",
				creator:    "UFRED",
			},
			cArgs: commandArgs{
				gameid:  "game1",
				creator: "UFRED",
			},
		},
		testArgs{name: "bad Slack ID",
			wantErr: true,
			errText: "OnGameStarted: A valid Slack ID",
			gArgs: gameArgs{
				gameid:     "game1",
				creator:    "UNOMATTER",
			},
			cArgs: commandArgs{
				gameid:  "game1",
				creator: "I_no_valido",
			},
		},
		testArgs{name: "missing game ID",
			wantErr: true,
			errText: "OnGameStarted: (mock) not a valid Game ID",
			gArgs: gameArgs{
				gameid:     "game1",
				creator:    "UNOMATTER",
			},
			cArgs: commandArgs{
				gameid:  "",
				creator: "UCREATE",
			},
			gPoolCtrl: gPoolControls{
				startGameErr: "(mock) not a valid Game ID",
			},
		},
		testArgs{name: "GamePool returns an error",
			wantErr: true,
			errText: "mock GamePool error message",
			gArgs: gameArgs{
				gameid:  "error me",
				creator: "USOMEONE",
			},
			cArgs: commandArgs{
				gameid:  "error me",
				creator: "USOMEONE",
			},
			gPoolCtrl: gPoolControls{
				startGameErr: "mock GamePool error message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGame := newGameFromArgs(tt.gArgs)
			gPool.AddGame(testGame)
			setMongoControlsFromArgs(mongo, tt.mongoCtrl)
			setGPoolControlsFromArgs(gPool, tt.gPoolCtrl)
			err := testHandler.OnGameStarted(tt.cArgs.gameid, tt.cArgs.creator)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), "OnGameStarted:", "All errors should start with the func name", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
			}
		})
	}
}

func TestHandler_GetGameStatus(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, mongo, "Placeholder to use mongo mock -- remove if mocking not needed")
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	testGame := newGameFromArgs(
		gameArgs{
			gameid:     "statusChecker",
			creator:    "USOMEONE",
			numPlayers: 7,
		},
	)
	gPool.AddGame(testGame)
	t.Run("positive", func(t *testing.T) {
		statusReport, exists := testHandler.GetGameStatus("statusChecker")
		require.True(t, exists, "Positive test should say the report exists")
		require.NotNil(t, statusReport, "Status report should exist")
	})
	t.Run("missing game ID", func(t *testing.T) {
		setGPoolControlsFromArgs(gPool, gPoolControls{
			getGameErr: "(mock) missing ID",
		})
		statusReport, exists := testHandler.GetGameStatus("notme")
		require.False(t, exists, "Negative test should say the report doesn't exist")
		require.Equal(t, "", statusReport, "Status report should exist")
	})
}

func TestHandler_GetGamesList(t *testing.T) {
	testHandler, mongo, gPool, blog := getHandlerWithMocksAndLogger(t)
	require.NotNil(t, mongo, "Placeholder to use mongo mock -- remove if mocking not needed")
	require.NotNil(t, blog, "Placeholder to use blog -- remove when log validation added")
	testGames := []*types.Game{
		newGameFromArgs(
			gameArgs{
				gameid:     "list_fodder_1",
				creator:    "USOMEONE",
				numPlayers: 1,
			},
		),
		newGameFromArgs(
			gameArgs{
				gameid:     "list_fodder_2",
				creator:    "USOMEONE",
				numPlayers: 2,
			},
		),
	}
	gPool.GamesToReturn = testGames

	t.Run("positive", func(t *testing.T) {
		gamesList := testHandler.GetGamesList()
		require.NotNil(t, gamesList, "Games List should exist")
		require.Contains(t, gamesList, "<h2>Games List</h2>")
		require.Contains(t, gamesList, " timestamp: ")
		require.Contains(t, gamesList, "list_fodder_1")
		require.Contains(t, gamesList, "list_fodder_2")
	})
}

/*** Helpers ***/

func getHandlerWithMocksAndLogger(t *testing.T) (testHandler *Handler, mockMongo *dao.MockMongoSession, mockGPool *types.MockGamePool, logBuf *bytes.Buffer) {
	mockMongo = dao.NewMockMongoSession()
	mockGames := []*types.Game{
		&types.Game{ID: "mockity", GameCreator: "UGOD"},
	}
	mockGPool = &types.MockGamePool{
		GamesToReturn: mockGames,
	}
	logBuf = &bytes.Buffer{}
	logLabel := "handler_test: "
	blog := log.New(logBuf, logLabel, 0)
	testHandler = NewHandler(mockGPool, mockMongo, blog)
	require.NotNil(t, testHandler, "Make sure creation worked")
	require.NotNil(t, testHandler.gPool, "Make sure we have a valid GamePool")
	return
}

// newGameFromArgs builds a Game for test purposes using a mix of required and optional fields with default values.
// gameid		string // required
// creator 		string // required
// killdict 	string // default: afile.txt
// passcode 	string // default: melod
// numPlayers	int    // required
// status  		types.GameStatus // default: Starting
func newGameFromArgs(args gameArgs) *types.Game {
	if args.killdict == "" {
		args.killdict = "afile.txt"
	}
	if args.passcode == "" {
		args.passcode = "melod"
	}

	gce, _ := events.NewGameCreatedEvent(args.gameid, slack.NewInline(args.creator), args.killdict, args.passcode)
	myGame := types.NewGameFromEvent(gce)
	myGame.StartPlayers = args.numPlayers

	return &myGame
}

func setGPoolControlsFromArgs(gpool *types.MockGamePool, args gPoolControls) {
	gpool.AddGameError = args.addGameErr
	gpool.AddPlayerError = args.addPlayerErr
	gpool.CanAddError = args.canAddErr
	gpool.GetGameError = args.getGameErr
	gpool.StartGameError = args.startGameErr
	gpool.GamesToReturn = args.gamesList
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
