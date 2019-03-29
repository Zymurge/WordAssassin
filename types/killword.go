package types

import (
	"fmt"

	bson "go.mongodb.org/mongo-driver/bson"
)

const (
	// MinCharLength is the minimum number of characters allowed in a valid word
	MinCharLength = 4
)

// KillWord is a data structure used to persist dictionary words
type KillWord struct {
	ID          string        `json:"id" bson:"_id"`
	DictID		string
	Word		string
}

// NewKillWord creates a new instance of a validated KillWord
func NewKillWord(dictID, word string) (KillWord, error) {
	//TODO: validate inputs: valid dictID name and word qualifiers	
	id := fmt.Sprintf("%s+%s", dictID, word)
	return KillWord{id, dictID, word}, nil
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