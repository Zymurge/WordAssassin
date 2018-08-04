package main

import (
	"wordassassin/types/events"
	"testing"
	"github.com/stretchr/testify/require"
	"time"

	types "wordassassin/types"
	dao "wordassassin/persistence"
)

// need to move this to a centralized place for tests
const (
	TestMongoURL   string = "localhost:27017"
	TestDbName     string = "testDB"
	TestCollection string = "TestCollection"
)

type mockControls struct {
	connectMode string
	writeMode string
	returnVal interface{}
}

func TestHandler_OnPlayerAdded(t *testing.T) {
	type args struct {
		gameid	string
		slackid string
		name    string
		email   string
	}

	mongo := dao.MockMongoSession{}
	testGPool := types.GamePool{}
	testPPool := types.PlayerPool{}
	testHandler := NewHandler(&testGPool, &testPPool, &mongo)

	tests := []struct {
		name    string
		h       Handler
		wantErr bool
		errText	string
		args    args
		mock	mockControls
	}{
		{ "positive", testHandler, false, "",
			args{ "game1", "@fred", "fred", "fred@bedrock.org" }, 
			mockControls{"positive", "positive", nil} },
		{ "PlayerPool error (from dup)", testHandler, true, "out of sync",
			args{ "game1", "@dupe_player", "dupey", "thesameguy@some.org" }, 
			mockControls{"positive", "positive", nil} },
		{ "duplicate slackID (at mongo)", testHandler, true, "@fred already added",
			args{ "game1", "@fred", "fred", "fred@bedrock.org" }, 
			mockControls{"positive", "duplicate", nil} },
		{ "empty slackID argument", testHandler, true, "missing SlackID",
			args{ "game1", "", "whatev", "bad@email.org" }, 
			mockControls{"positive", "positive", nil} },
		{ "gameID doesn't exist", testHandler, true, "missing_gameid doesn't exist",
			args{ "missing_gameid", "@someone", "someone", "bad@email.org" }, 
			mockControls{"positive", "positive", nil} },
		{ "gameID already started", testHandler, true, "not accepting players",
			args{ "started_gameid", "@toolate", "lagger", "tomorrow@procrastinate.org" }, 
			mockControls{"positive", "positive", nil} },
		{ "empty GameID argument", testHandler, true, "doesn't exist",
			args{ "", "@someone", "someone","bad@email.org" }, 
			mockControls{"positive", "positive", nil} },
		{ "mongo fail", testHandler, true, "connect",
			args{ "game1", "n/a", "n/a", "bad@email.org" }, 
			mockControls{"no connect", "positive", nil} },
	}

	// Add a game named "game1" for cases that expect it
	game1 := types.Game{ 
		ID: "game1", 
		TimeCreated: time.Now(), 
		GameCreator: "testmaster", 
		KillDictionary: "websters", 
		Status: types.Starting,
	}
	testGPool.AddGame(&game1)
	// Add a game named "started_gameid" for cases that expect it
	started := types.Game{ 
		ID: "started_gameid", 
		TimeCreated: time.Now(), 
		GameCreator: "testmaster", 
		KillDictionary: "websters", 
		Status: types.Playing,
	}
	testGPool.AddGame(&started)
	// Add a preexisting player to support duplicate cases
	dupEv, _ := events.NewPlayerAddedEvent("game1","@dupe_player", "", "")
	dupPlayer := types.NewPlayerFromEvent(dupEv)
	testPPool.AddPlayer(&dupPlayer)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongo.ConnectMode = tt.mock.connectMode
			mongo.WriteMode = tt.mock.writeMode
			err := tt.h.OnPlayerAdded(tt.args.gameid, tt.args.slackid, tt.args.name, tt.args.email)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual,err := testPPool.GetPlayer(tt.args.gameid, tt.args.slackid)
				require.NoErrorf(t, err, "Didn't want to see this: %v", err)
				require.Equal(t, tt.args.name, actual.Name, "Didn't find player added despite success")
			}
		})
	}
}

func TestHandler_OnGameCreated(t *testing.T) {
	type args struct {
		gameid    string
		creator   string
		killdict  string
		passcode  string
	}

	mongo := dao.MockMongoSession{}
	testGPool := types.GamePool{}
	testPPool := types.PlayerPool{}
	testHandler := NewHandler(&testGPool, &testPPool, &mongo)

	tests := []struct {
		name    string
		h       Handler
		wantErr bool
		errText	string
		args    args
		mock	mockControls
	}{
		{ "positive", testHandler, false, "",
			args{ "game1", "@fred", "somedict.txt", "topsecret" }, 
			mockControls{"positive", "positive", nil} },
		{ "GamePool error (from dup)", testHandler, true, "out of sync",
			args{ "dupe_game", "@testmaster", "somedict.txt", "topsecret" }, 
			mockControls{"positive", "positive", nil} },
		{ "duplicate gameID (at mongo)", testHandler, true, "dupe_game already created",
			args{ "dupe_game", "@testmaster", "somedict.txt", "topsecret" }, 
			mockControls{"positive", "duplicate", nil} },
		{ "empty gameID argument", testHandler, true, "missing GameID",
			args{ "", "@someone", "whatev", "xxx" }, 
			mockControls{"positive", "positive", nil} },
		{ "mongo fail", testHandler, true, "connect",
			args{ "fail", "n/a", "n/a","bad@email.org"}, 
			mockControls{"no connect", "positive", nil} },
	}

	// Add a game for tests that need a pre-existing id
	dupGame := types.Game{ 
		ID: "dupe_game", 
		TimeCreated: time.Now(), 
		GameCreator: "@testmaster", 
		KillDictionary: "websters", 
		Status: types.Starting,
	}
	testGPool.AddGame(&dupGame)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongo.ConnectMode = tt.mock.connectMode
			mongo.WriteMode = tt.mock.writeMode
			err := tt.h.OnGameCreated(tt.args.gameid, tt.args.creator, tt.args.killdict, tt.args.passcode)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual,exists := testGPool.GetGame(tt.args.gameid)
				require.NotNilf(t, exists, "Expected % to exist in the gamepool", tt.args.gameid)
				require.Equal(t, tt.args.gameid, actual.GetID(), "Didn't find game added despite success")
			}
		})
	}
}
