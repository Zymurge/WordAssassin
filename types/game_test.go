package types

import (
	"testing"
	"github.com/stretchr/testify/require"
	"time"

	events "wordassassin/types/events"
)

func TestGamePool(t *testing.T) {
	// Setup
	target := GamePool{}
	require.NotNil(t, target)

	g1 := NewGameFromEvent( events.GameCreatedEvent{
			ID:	         "g1",
			TimeCreated: time.Now(),
			GameCreator: "some test code",
			KillDictionary: "a file somewhere",
		} )
	if err := target.AddGame(&g1); err != nil {
		require.NoErrorf(t, err, "Didn't want to see: %v", err)
	}
	g2 := NewGameFromEvent( events.GameCreatedEvent{
			ID:	         "g2",
			TimeCreated: time.Now(),
			GameCreator: "that same code",
			KillDictionary: "a file somewhere else",
		} )
	if err := target.AddGame(&g2); err != nil {
		require.NoErrorf(t, err, "Didn't want to see: %v", err)
	}
}