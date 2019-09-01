package types

import (
	"fmt"

	mongo "wordassassin/persistence"
)

// KillDictionary represents a collection of valid words to use within a game of wordassassin
type KillDictionary struct {
	mongo mongo.MongoAbstraction
	ID    string
	words []string
}

const (
	// CollectionName is the name of the mongo collection where words are stored
	CollectionName = "killwords"
)

// NewKillDictionary creates an unique instance
// Unique ID enforced by persisted
// Input list is scrubbed to allow valid values only (single word, more than 4 letters)
func NewKillDictionary(m mongo.MongoAbstraction, id string, words ...string) KillDictionary {
	dict := KillDictionary{m, id, make([]string, 0)}
	// TODO: validate each word and remove the bad ones
	// TODO: write each word throough the AddWord method, to leverage scrubbing rules
	for _, word := range words {
		if err := dict.AddWord(word); err != nil {
			// eat errors for now, since it only indicates an illegal word
		}
	}
	return dict
}

// AddWord adds a new word to the dictionary
// Filters out words that do not meet the acceptable criteria:
// - Word must be 4 or more characters
// Returns an error on unsuccessful addition
func (kd *KillDictionary) AddWord(word string) error {
	// create a mongo friendly object to persist. Validate kw format as a side effect.
	kw, err := NewKillWord(kd.ID, word)
	if err != nil {
		return fmt.Errorf("AddWord: %v", err)
	}

	// attempt mongo write
	if err = kd.mongo.WriteCollection(CollectionName, &kw); err != nil {
		return err
	}
	// add to kd array
	kd.words = append(kd.words, word)
	// fallthrough success
	return nil
}

// Count returns the number of available words in the dictionary
func (kd *KillDictionary) Count() int {
	return len(kd.words)
}

// GetKillWord selects a word at random from the dictionary
func (kd *KillDictionary) GetKillWord() string {
	panic("GetKillWord - Not implemented")
}

// RestoreFromMongo finds the dictionary
// func (kd *KillDictionary) RestoreFromMongo() string {
// 	panic("Not implemented")
// }
