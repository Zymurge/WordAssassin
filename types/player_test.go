package types

import (
	"testing"
	"github.com/stretchr/testify/require"

	events "wordassassin/types/events"
)

func TestPlayerCreation(t *testing.T) {
	expectedGameID := "the_jungle"
	expectedSlackID := "Ubigape"
	expectedName :=  "King Kong"
	expectedEmail := "kk@jung.le"
	expectedID := expectedGameID + "+" + expectedSlackID
	actual := NewPlayerFromEvent( events.NewPlayerAddedInline(expectedGameID, expectedSlackID, expectedName, expectedEmail) )
	require.NotNil(t, actual)
	require.Equal(t, expectedID, actual.GetID())
	require.Equal(t, expectedName, actual.Name)
	require.Equal(t, expectedEmail, actual.Email)
	require.Equal(t, "", actual.Target, "Should be created with target unset")
	require.Equal(t, "", actual.KillWord, "Should be created with killword unset")
}

func TestPlayer_SetTarget(t *testing.T) {
	actual := NewPlayerFromEvent( events.NewPlayerAddedInline("a_game", "Uplayer", "The Big P", "playuh@game.org") )
	require.NotNil(t, actual)
	expectedTarget := "Usonofab"
	expectedKillword := "MySharona"
	actual.SetTarget(expectedTarget, expectedKillword)
	require.Equal(t, expectedTarget, actual.Target)
	require.Equal(t, expectedKillword, actual.KillWord)
}
