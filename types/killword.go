package types

import (
	"fmt"

	bson "go.mongodb.org/mongo-driver/bson"
)

// KillWord is a data structure used to persist dictionary words
type KillWord struct {
	ID     string `json:"id" bson:"_id"`
	DictID string
	Word   string
}

const (
	// KillWordMinCharLength is the minimum number of characters allowed in a valid word
	KillWordMinCharLength int = 4
)

// NewKillWord creates a new instance of a validated KillWord
func NewKillWord(dictID, word string) (response KillWord, err error) {
	// validate word (length only so far)
	if len(word) < KillWordMinCharLength {
		err = fmt.Errorf("%s does not meet the minimum char length %d", word, KillWordMinCharLength)
		return
	}
	// validate KillDictionary (non-empty string for now)
	if len(dictID) < 1 {
		err = fmt.Errorf("A blank ID is not a valid KillDictionary")
		return
	}
	id := fmt.Sprintf("%s+%s", dictID, word)
	response = KillWord{id, dictID, word}
	return
}

// Decode populates this instance from the supplied bson
func (kw *KillWord) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, kw); err != nil {
		return err
	}
	return nil
}

// GetID getter for ID field
func (kw *KillWord) GetID() string {
	return kw.ID
}
