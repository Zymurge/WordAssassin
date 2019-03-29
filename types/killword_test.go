package types

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestKillWord_positive(t *testing.T) {
	d := "myDict"
	w := "aValidWord"
	expectedID := d + "+" + w
	actual, err := NewKillWord(d, w)
	require.NoError(t, err, "Shouldn't throw on success")
	require.NotNil(t, actual, "Want a usable instance back")
	require.Equal(t, d, actual.DictID)
	require.Equal(t, w, actual.Word)
	require.Equal(t, expectedID, actual.GetID())
}

func TestKillWord_mimimum_length(t *testing.T) {
	d := "myDict"
	w := "no"
	actual, err := NewKillWord(d, w)
	require.Error(t, err, "Should throw an error")
	require.Nil(t, actual, "Don't want a usable instance back")
	require.Contains(t, "minimum", err.Error())
}

func TestKillWord_illegal_dict(t *testing.T) {
	d := ""
	w := "goodword"
	actual, err := NewKillWord(d, w)
	require.Error(t, err, "Should throw an error")
	require.Nil(t, actual, "Don't want a usable instance back")
	require.Contains(t, "dictionary", err.Error())
}