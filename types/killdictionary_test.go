package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	dao "wordassassin/persistence"
)

func TestKillDictionary_NewKillDictionary(t *testing.T) {
	mockMongo := dao.NewMockMongoSession()
	wordList := []string{
		"yohoho",
		"tiddlywinks",
		"crunchberries",
	}
	type args struct {
		id   string
		word []string
	}
	tests := []struct {
		name string
		args args
		want KillDictionary
	}{
		{name: "Positive",
			args: args{
				id:   "kd1",
				word: wordList,
			},
			want: KillDictionary{mockMongo, "kd1", wordList},
		},
		{name: "Filter short words",
			args: args{
				id:   "filterme",
				word: []string{"valid1", "valid2", "no", "valid3"},
			},
			want: KillDictionary{mockMongo, "filterme", []string{"valid1", "valid2", "valid3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKillDictionary(mockMongo, tt.args.id, tt.args.word...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KillDictionary.NewKillDictionary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKillDictionary_AddWord(t *testing.T) {
	mockMongo := dao.NewMockMongoSession()
	target := NewKillDictionary(mockMongo, "target1", "pre-existing")
	type args struct {
		word string
	}
	tests := []struct {
		name      string
		args      args
		err       bool
		msg       string
		mongoCtrl dao.MongoControls // default mock controls
	}{
		{name: "Positive",
			args: args{word: "good-word"},
			err:  false,
		},
		{name: "Short word",
			args: args{word: "xx"},
			err:  true,
			msg:  "minimum",
		},
		{name: "Duplicate",
			args: args{word: "pre-existing"},
			err:  true,
			msg:  "duplicate",
			mongoCtrl: dao.MongoControls{
				ConnectMode: "positive",
				WriteMode:   "duplicate",
				ReturnVal:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMongo.SetMongoControlsFromArgs(tt.mongoCtrl)
			got := target.AddWord(tt.args.word)
			if tt.err {
				require.NotNil(t, got, "Expected error, didn't get one")
				require.Contains(t, got.Error(), tt.msg, "Err msg did not contain required keyword")
			} else {
				require.Nil(t, got, "Did not want any error on return")
			}
		})
	}
}
