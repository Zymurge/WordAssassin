package main

import (
	"fmt"
	"strings"
	"time"
	"log"

	persistence "wordassassin/persistence"
	types "wordassassin/types"
	events "wordassassin/types/events"
	slack "wordassassin/slack"
)

// Handler contains the context necessary to process events and put everything where it belongs. Needs to be aware
// of persistence, the game pool, the player pool, etc
type Handler struct {
	gPool	 types.GamePoolAbstraction
	mongo 	 persistence.MongoAbstraction
	logger   *log.Logger
	// other stuff?
}

// NewHandler creates a handler instance using the injected dependencies (hint, hint: they're for testing)
func NewHandler(gp types.GamePoolAbstraction, m persistence.MongoAbstraction, l *log.Logger) (h *Handler) {
	if gp == nil {
		panic("GamePool argument is nil")
	}
	if m == nil {
		panic("MongoSession argument is nil")
	}
	if l == nil {
		panic("Logger argument is nil")
	}
	h = &Handler{
		gPool: gp,
		mongo: m,
		logger: l,
	}
	l.Printf("Startup: Handler created")
	return
}

// OnGameCreated handles coordination when a game is created for this server.
// -- An event is created and persisted to mongo
// -- The new game is added to the game pool
// Errors:
// -- validation errors on all params
// -- duplicate game created (GameID already exists)
// -- mongo issue
func (h Handler) OnGameCreated(gameid, creator, killdict, passcode string) (err error) {
	creatorID, err := slack.New(creator)
	if err != nil {
		return fmt.Errorf("OnGameCreated: %v", err)
	}
	// Create and persist the event to request a new game
	var ev events.GameCreatedEvent
	if ev, err = events.NewGameCreatedEvent(gameid, creatorID, killdict, passcode); err != nil {
		err = fmt.Errorf("OnGameCreated: %v", err)
		return
	}
	if mongoerr := h.mongo.WriteCollection("events", &ev); mongoerr != nil {
		// Want to handle errors with more graceful wording for downstream consumers
		if strings.Contains(mongoerr.Error(), "duplicate") {
			err = fmt.Errorf("OnGameCreated: Game %s already created", gameid)
		} else {
			err = fmt.Errorf("OnGameCreated: Mongodb issue on GameCreated event write: %v", mongoerr)
		}
		return
	}

	// Create and register the game object in the game pool
	game := types.NewGameFromEvent(ev)
	if gperr := h.gPool.AddGame(&game); gperr != nil {
		// Should catch all dups at the event level
		if strings.Contains(gperr.Error(), "duplicate") {
			err = fmt.Errorf("OnGameCreated: Something bad happened. GamePool out of sync with mongo events")
			return
		}
		err = fmt.Errorf("OnGameCreated: Issue on GameCreated add to GamePool: %v", gperr)
		return
	}

	return nil
}

// OnGameStarted handles activiting a game from the starting stage into playing.
// Only the original game creator is allowed to start a given gameid.
// Errors:
// -- gameid empty
// -- valid slackid
// -- gameid does not exists 
// -- gameid not in 'starting' state
// -- slackid does not match the creating slackid
func (h *Handler) OnGameStarted(gameid string, creator string) (err error) {
	// First, make sure there's already a game and it's not started yet
	creatorID, err := slack.New(creator)
	if err != nil {
		return fmt.Errorf("OnGameStarted: %v", err)
	}
	err = h.gPool.StartGame(gameid, creatorID)
	if err != nil {
		return fmt.Errorf("OnGameStarted: %v", err)
	}
	return
}

// OnPlayerAdded handles coordination when a player is added to the game:
// -- A unique player ID is created from the combo of gameid and slackid
// -- An event is created and persisted to mongo
// -- The new player is added to the player pool
// Errors:
// -- gameid does not exists 
// -- gameid not in 'starting' state
// -- slackid empty
// -- duplicate player added
// -- mongo issue
// -- gamepool issue
func (h Handler) OnPlayerAdded(gameid string, slackid string, name string, email string) (err error) {
	// First, make sure there's already a game and it's accepting players
	if accepting, acceptErr := h.gPool.CanAddPlayers(gameid); !accepting {
		err = fmt.Errorf("OnPlayerAdded: game %s: %v", gameid, acceptErr)
		return
	}

	// Create and persist an event, unless it's a dupe. Rely on PlayerAddEvent ctor to validate inputs
	var ev events.PlayerAddedEvent
	if ev, err = events.NewPlayerAddedEvent(gameid, slackid, name, email); err != nil {
		err = fmt.Errorf("OnPlayerAdded: %v", err)
		return
	}

	if mongoerr := h.mongo.WriteCollection("events", &ev); mongoerr != nil {
		// Want to handle a dup write with more graceful wording for downstream consumers
		if strings.Contains(mongoerr.Error(), "duplicate") {
			err = fmt.Errorf("OnPlayerAdded: Player %s already added to game %s", slackid, gameid)
		} else {
			err = fmt.Errorf("OnPlayerAdded: Mongodb write issue: %v", mongoerr)
		}
		return
	}

	if gpErr := h.gPool.AddPlayerToGame(gameid, ev); gpErr != nil {
		err = fmt.Errorf("OnPlayerAdded: %v", gpErr)
	}
	return
}

// GetGameStatus produces a game status report for the specified 
// Provides an existence check in lieu of error messages
func (h *Handler) GetGameStatus(gameid string) (result string, exists bool) {
	var game *types.Game
	if game, exists = h.gPool.GetGame(gameid); !exists {
		return
	}
	result = game.GetStatusReport()
	return
}

// GetGamesList provides a listing of all of the games in the GamePool
func (h *Handler) GetGamesList() (result string) {
	result = "<h2>Games List</h2>\n"
	result += "  timestamp: " + time.Now().String() + "\n<p>\n"
	games := h.gPool.GetGamesList()
	for _, v := range games {
		line := fmt.Sprintf("<li>%s: %s, %d players</li>", v.GetID(), v.GetStatus(), v.StartPlayers)
		result += line
	}
	return
}
