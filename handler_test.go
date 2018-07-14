package wordassassin

import (
	"testing"
	"github.com/stretchr/testify/require"

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
		name    string
		slackid string
		email   string
	}

	testPool := types.PlayerPool{ GameID: "testGameID" }
	mongo := dao.MockMongoSession{}
	testHandler := NewHandler(&testPool, &mongo)

	tests := []struct {
		name    string
		h       Handler
		wantErr bool
		errText	string
		args    args
		mock	mockControls
	}{
		{ "positive", testHandler, false, "",
			args{ "fred", "@fred","fred@bedrock.org"}, 
			mockControls{"positive", "positive", nil} },
		{ "PlayerPool error (from dup)", testHandler, true, "duplicate",
			args{ "fred", "@fred","fred@bedrock.org"}, 
			mockControls{"positive", "positive", nil} },
		{ "duplicate ID (at mongo)", testHandler, true, "duplicate",
			args{ "fred", "@fred","fred@bedrock.org"}, 
			mockControls{"positive", "duplicate", nil} },
		{ "missing ID", testHandler, true, "missing",
			args{ "someone", "","bad@email.org"}, 
			mockControls{"positive", "positive", nil} },
		{ "mongo fail", testHandler, true, "connect",
			args{ "n/a", "n/a","bad@email.org"}, 
			mockControls{"no connect", "positive", nil} },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongo.ConnectMode = tt.mock.connectMode
			mongo.WriteMode = tt.mock.writeMode
			err := tt.h.OnPlayerAdded(tt.args.name, tt.args.slackid, tt.args.email)
			if tt.wantErr {
				require.Errorf(t, err, "Was looking for an error containing '%s' but got none", tt.errText)
				require.Contains(t, err.Error(), tt.errText, "Got an error but didn't find '%s' in the content", tt.errText)
			} else {
				require.NoErrorf(t, err, "Was expecting successful call, but got err: %v", err)
				actual,err := testPool.GetPlayer(tt.args.slackid)
				require.NoErrorf(t, err, "Didn't want to see this: %v", err)
				require.Equal(t, tt.args.name, actual.Name, "Didn't find player added despite success")
			}
		})
	}
}
