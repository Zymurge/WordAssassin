package main

import (
	"fmt"
	"strings"

	persistence "wordassassin/persistence"
	types "wordassassin/types"
	events "wordassassin/types/events"
)

// Handler contains the context necessary to process events and put everything where it belongs. Needs to be aware
// of persistence, the game pool, the player pool, etc
type Handler struct {
	gPool  *types.GamePool
	pPool  *types.PlayerPool
	mongo persistence.MongoAbstraction
	// other stuff?
}

// NewHandler creates a handler instance using the injected dependencies (hint, hint: they're for testing)
func NewHandler(gg *types.GamePool, pp *types.PlayerPool, m persistence.MongoAbstraction) Handler {
	return Handler{
		gPool: gg,
		pPool: pp,
		mongo: m,
	}
}

// OnGameCreated handles coordination when a game is created for this server.
// A unique game ID is created here.
// Errors:
// -- creator or killdict is empty
func (h Handler) OnGameCreated(gameid, creator, killdict, passcode string) (err error) {
	// Validate the params

	// Create and persist the event to request a new game
	var ev events.GameCreatedEvent
	if ev, err = events.NewGameCreatedEvent(gameid, creator, killdict, passcode); err != nil {
		 return
	}
	if mongoerr := h.mongo.WriteCollection("events", &ev); mongoerr != nil {
		// Want to handle a dup write with more graceful wording for downstream consumers
		if strings.Contains(mongoerr.Error(), "duplicate") {
			err = fmt.Errorf("Game %s already created", gameid)
			return
		}
		err = fmt.Errorf("Mongodb issue on GameCreated event write: %s", mongoerr.Error())
		return
	}

	// Create and register the game object in the game pool
	game := types.NewGameFromEvent(ev)
	if gperr := h.gPool.AddGame(&game); gperr != nil {
		// Should catch all dups at the event level
		if strings.Contains(gperr.Error(), "duplicate") {
			// TODO: log this issue
			err = fmt.Errorf("Something bad happened. GamePool out of sync with mongo events")
			return
		}
		err = fmt.Errorf("Issue on GameCreated add to GamePool: %s", gperr.Error())
		return
	}

	return nil
}

// OnPlayerAdded handles coordination when a player is added to the game.
// A unique player ID is created from the combo of gameid and slackid.
// Errors:
// -- gameid or slackid empty
// -- gameid not exists and in 'starting' state
// -- duplicate player added
func (h Handler) OnPlayerAdded(gameid string, slackid string, name string, email string) (err error) {
	// First, make sure there's already a game and it's accepting players
	game, exists := h.gPool.GetGame(gameid)
	if !exists {
		err = fmt.Errorf("The requested GameID: %s doesn't exist on this server", gameid)
		return
	}
	if game.Status != types.Starting {
		err = fmt.Errorf("The requested GameID: %s is not accepting players. State=%d", gameid, game.Status)
		return
	}

	// Create and persist and event, unless it's a dupe. Rely on PlayerAddEvent ctor to validate inputs
	var ev events.PlayerAddedEvent
	if ev, err = events.NewPlayerAddedEvent(gameid, slackid, name, email); err != nil {
		 return
	}
	if mongoerr := h.mongo.WriteCollection("events", &ev); mongoerr != nil {
		// Want to handle a dup write with more graceful wording for downstream consumers
		if strings.Contains(mongoerr.Error(), "duplicate") {
			err = fmt.Errorf("Player %s already added to game %s", slackid, gameid)
			return
		}
		err = fmt.Errorf("Mongodb issue on AddPlayer event write: %s", mongoerr.Error())
		return
	}

	// Create the Player instance and add to the PlayerPool
	player := types.NewPlayerFromEvent(ev)
	if hperr := h.pPool.AddPlayer(&player); hperr != nil {
		// Should catch all dups at the event level
		if strings.Contains(hperr.Error(), "duplicate") {
			// TODO: log this issue
			err = fmt.Errorf("Something bad happened. PlayerPool out of sync with mongo events")
			return
		}
		err = fmt.Errorf("Issue on AddPlayer add to PlayerPool: %s", hperr.Error())
		return
	}
	// Persist the pPool, or should it auto persist on state change?

	return nil
}
