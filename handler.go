package main

import (
	"strings"
	"fmt"
	"time"
	persistence "wordassassin/persistence"
	types "wordassassin/types"
	events "wordassassin/types/events"
)

// Handler contains the context necessary to process events and put everything where it belongs. Needs to be aware
// of persistence, the player pool, etc
type Handler struct {
	pool  *types.PlayerPool
	mongo persistence.MongoAbstraction
	// other stuff?
}

// NewHandler creates a handler instance using the injected dependencies (hint, hint: they're for testing)
func NewHandler(p *types.PlayerPool, m persistence.MongoAbstraction) Handler {
	return Handler{
		pool:  p,
		mongo: m,
	}
}

// OnPlayerAdded handles coordination when a player is added to the game
func (h Handler) OnPlayerAdded(name string, slackid string, email string) error {
	// Parse and validate input
	// Generate and persist a PlayerAddedEvent
	if slackid == "" {
		return fmt.Errorf("missing ID field")
	}
	ev := events.PlayerAddedEvent{
		ID:          slackid, // assume slackid becomes the unique identifier
		Name:        name,
		SlackID:     slackid,
		Email:       email,
		TimeCreated: time.Now(),
	}
	if err := h.mongo.WriteCollection("events", &ev); err != nil {
		// Want to handle a dup write with more graceful wording for downstream consumers
		if strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("Player %s already added to game %s", slackid, h.pool.GameID)
		}
		return err
	}
	// Create the Player instance
	player := types.NewPlayerFromEvent(ev)
	// Add to the PlayerPool
	if err := h.pool.AddPlayer(&player); err != nil {
		// Should catch all dups at the event level
		if strings.Contains(err.Error(), "duplicate") {
			// TODO: log this issue
			return fmt.Errorf("Something bad happened. PlayerPool out of sync with mongo events")
		}
		return err
	}
	// Persist the pool, or should it auto persist on state change?

	return nil
}
