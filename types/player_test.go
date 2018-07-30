package types

import (
	"testing"
	"github.com/stretchr/testify/require"

	events "wordassassin/types/events"
)

func TestPlayerPool(t *testing.T) {
	// Setup
	target := PlayerPool{}
	require.NotNil(t, target)
	pe, err := events.NewPlayerAddedEvent("game1", "@Joe","Joe", "joe@wa.org" )
	require.NoErrorf(t, err, "Error on NewPlayerAddedEvent creation: %v", err)
	p1 := NewPlayerFromEvent(pe)
	if err := target.AddPlayer(&p1); err != nil {
		require.NoErrorf(t, err, "Didn't want to see: %v", err)
	}
	pe, err = events.NewPlayerAddedEvent("game1", "@Jim","Jim", "jim@wa.org" )
	require.NoErrorf(t, err, "Error on NewPlayerAddedEvent creation: %v", err)
	p2 := NewPlayerFromEvent(pe)
	if err := target.AddPlayer(&p2); err != nil {
		require.NoErrorf(t, err, "Didn't want to see: %v", err)
	}

	// Execute
	t.Run("Add: Duplicate ID", func(t *testing.T) {
		err := target.AddPlayer(&p1)
		require.Error(t, err, "Should have thrown on duplicate")
		require.Contains(t, err.Error(), "duplicate", "Error must mention the issue")
	})
	t.Run("Add: Missing ID", func(t *testing.T) {
		p3 := p1
		p3.ID = ""
		err := target.AddPlayer(&p3)
		require.Error(t, err, "Should have thrown on missing ID")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetByID: positive", func(t *testing.T) {
		actual, err := target.GetPlayerByID(p2.GetID())
		require.NoErrorf(t, err, "Positive tosses no errors, but this one did!: %v", err)
		require.Equal(t, &p2, actual)
	})
	t.Run("GetByID: missing ID", func(t *testing.T) {
		_, err := target.GetPlayerByID("bad ID")
		require.Errorf(t, err, "Error you will If ID you miss")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("GetByID: before players initialized", func(t *testing.T) {
		badTarget := PlayerPool{}
		_, err := badTarget.GetPlayerByID("who cares?")
		require.Errorf(t, err, "Not sure what happens here")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})
	t.Run("Get: positive", func(t *testing.T) {
		actual, err := target.GetPlayer(p2.GameID, p2.SlackID)
		require.NoErrorf(t, err, "Positive tosses no errors, but this one did!: %v", err)
		require.Equal(t, &p2, actual)
	})
	t.Run("Get: missing ID", func(t *testing.T) {
		_, err := target.GetPlayer(p2.GameID, "bad ID")
		require.Errorf(t, err, "Error you will If ID you miss")
		require.Contains(t, err.Error(), "missing", "Error must mention the issue")
	})

}