package types

import (
	"strconv"
	"time"
	//"fmt"
	//"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testGTstr string = "{'id': '123', 'game': 'testGame', 'status': 'running', 'starttime': 0, 'startplayers': 99, 'remainplayers': 33}"
)

var	testGT GameTracker = GameTracker{
		ID: "jsonform",
		Game: "some game",
		Status: Starting,
		StartTime: time.Now(),
		StartPlayers:  13,
		RemainPlayers: 3,
	}

func testGTjson(t *testing.T, gt GameTracker) []byte {
	marshalledJSONstarttime, err := gt.StartTime.MarshalJSON()
	require.NoErrorf(t, err, "Got %v in parsing time", err)
	testGTjson := make([]byte, 0)
	testGTjson = append( testGTjson, []byte(`{"id":"`)[:]...)
	testGTjson = append( testGTjson, gt.ID[:]...)
	testGTjson = append( testGTjson, []byte(`","game":"`)[:]...)
	testGTjson = append( testGTjson, gt.Game[:]...)
	testGTjson = append( testGTjson, []byte(`","status":`)[:]...)
	testGTjson = append( testGTjson, strconv.Itoa(int(gt.Status))[:]...)
	testGTjson = append( testGTjson, []byte(`,"starttime":`)[:]...)
	testGTjson = append( testGTjson, marshalledJSONstarttime[:]...)
	testGTjson = append( testGTjson, []byte(`,"startplayers":`)[:]...)
	testGTjson = append( testGTjson, strconv.Itoa(int(gt.StartPlayers))[:]...)
	testGTjson = append( testGTjson, []byte(`,"remainplayers":`)[:]...)
	testGTjson = append( testGTjson, strconv.Itoa(int(gt.RemainPlayers))[:]...)
	testGTjson = append( testGTjson, []byte(`}`)[:]...)
	return testGTjson
}

func TestStatusCtor(t *testing.T) {
	result := GameTracker{
		ID:           "123",
		Game:         "testGame",
		StartPlayers: 99,
	}
	require.IsType(t, GameTracker{}, result)
	require.Equal(t, "123", result.ID)
	require.Equal(t, "testGame", result.Game)
	require.Equal(t, 99, result.StartPlayers)
}

func TestGameTrackerFromJSON(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		result, err := GameTrackerFromJSON(testGTjson(t, testGT))
		require.NoError(t, err, "Positive test should not throw an error")
		require.IsType(t, GameTracker{}, result, "Should return a GameTracker struct")
		require.Equal(t, testGT.ID, result.ID, "ID not mapped as expected")
		require.Equal(t, testGT.Status, result.Status, "Status not mapped as expected")
		require.Equal(t, testGT.StartTime.Year(), result.StartTime.Year(), "Gotta get my Rush on here. Got time %s", result.StartTime.Year())
	})
	t.Run("Bad JSON format", func(t *testing.T) {
		testJSON := []byte(`{"id":"Me be bad","game":"broken json",badvariable":6,"status":"starting"}`)
		_, err := GameTrackerFromJSON(testJSON)
		require.Error(t, err, "Looking for an unmarshal error")
		require.Contains(t, err.Error(), "unmarshal", "Hope the unmarshal keyword is specified in the err message")
	})
	t.Run("Missing ID", func(t *testing.T) {
		testJSON := []byte(`{"id_poop":42,"status":` + strconv.Itoa(int(Finished)) + `}`)
		_, err := GameTrackerFromJSON(testJSON)
		require.Error(t, err, "Looking for a no ID error")
		require.Contains(t, err.Error(), "ID", "Hope the ID keyword is specified in the err message")
	})

}

func TestJSONForm(t *testing.T) {
	testGT := GameTracker{
		ID: "jsonform",
		Game: "some game",
		Status: Starting,
		StartTime: time.Now(),
		StartPlayers:  13,
		RemainPlayers: 3,
	}
	t.Run("Positive", func(t *testing.T) {
		expectedJSON := testGTjson(t,testGT)
		actualJSON := testGT.JSONForm()
		require.Equal(t, expectedJSON, actualJSON)
	})
}

/*
func TestLocFromString(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		result, err := LocFromString("3.6.9")
		require.NoError(t, err, "Positive test should not throw an error")
		assert.IsType(t, Loc{}, result, "Should return a Loc struct")
		assert.True(t,
			result.ID == "3.6.9" && result.X == 3 && result.Y == 6 && result.Z == 9,
			"X,Y,Z values not mapped as expected")
	})
	t.Run("Bad Delimiter", func(t *testing.T) {
		_, err := LocFromString("3.6*9")
		assert.Error(t, err, "Negative test should throw an error")
		assert.True(t, strings.Contains(err.Error(), "x.y.z"), "Expect an error message that mentions the proper format")
	})
}

// Table driven test to try a wide range of positive cases
func TestDistanceFrom(t *testing.T) {
	var cases = []struct {
		origin   Loc
		target   Loc
		expected int
	}{
		{newLoc(12, -7, 99), newLoc(19, 10, 99), 17},
		{newLoc(100, -7, 0), newLoc(113, 10, 99), 99},
		{newLoc(1, 2, 3), newLoc(-44, 2, -3), 45},
		{newLoc(0, 0, 0), newLoc(0, 0, 0), 0},
	}

	for num, c := range cases {
		t.Run(fmt.Sprintf("case#%d", num), func(t *testing.T) {
			//fmt.Printf( "-- case: %v.DistanceFrom( %v ) = %d\n", c.origin, c.target, c.expected )
			actual := c.origin.DistanceFrom(c.target)
			assert.Equal(t, c.expected, actual)
		})
	}
}

/*** Helpers ***

// Helper function to swallow the multiple return value. Allow a newLoc call within a struct declaration.
func newLoc(x int, y int, z int) Loc {
	result, _ := LocFromCoords(x, y, z)
	return result
}
*/
