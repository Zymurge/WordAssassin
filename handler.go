package main

import (
	"fmt"
	"strings"
	"time"
	"log"

	persistence "wordassassin/persistence"
	types "wordassassin/types"
	events "wordassassin/types/events"
)

// Handler contains the context necessary to process events and put everything where it belongs. Needs to be aware
// of persistence, the game pool, the player pool, etc
type Handler struct {
	gPool	 types.GamePoolAbstraction
	pPool 	 *types.PlayerPool
	mongo 	 persistence.MongoAbstraction
	logger   *log.Logger
	// other stuff?
}

// NewHandler creates a handler instance using the injected dependencies (hint, hint: they're for testing)
func NewHandler(gp types.GamePoolAbstraction, pp *types.PlayerPool, m persistence.MongoAbstraction, l *log.Logger) (h *Handler) {
	if gp == nil {
		panic("GamePool argument is nil")
	}
	if pp == nil {
		panic("PlayerPool argument is nil")
	}
	if m == nil {
		panic("MongoSession argument is nil")
	}
	if l == nil {
		panic("Logger argument is nil")
	}
	h = &Handler{
		gPool: gp,
		pPool: pp,
		mongo: m,
		logger: l,
	}
	l.Printf("Startup: Handler created")
	return
}

// OnGameCreated handles coordination when a game is created for this server.
// A unique game ID is created here.
// Errors:
// -- creator or killdict is empty
func (h Handler) OnGameCreated(gameid, creator, killdict, passcode string) (err error) {
	if gameid   == "" { return fmt.Errorf("OnGameCreated: cannot create game with blank gameid") }
	if creator  == "" { return fmt.Errorf("OnGameCreated: cannot create game with blank creator") }
	if killdict == "" { return fmt.Errorf("OnGameCreated: cannot create game with blank killdict") }
	if passcode == "" { return fmt.Errorf("OnGameCreated: cannot create game with blank passcode") }

	// Create and persist the event to request a new game
	var ev events.GameCreatedEvent
	if ev, err = events.NewGameCreatedEvent(gameid, creator, killdict, passcode); err != nil {
		return
	}
	if mongoerr := h.mongo.WriteCollection("events", &ev); mongoerr != nil {
		// Want to handle errors with more graceful wording for downstream consumers
		if strings.Contains(mongoerr.Error(), "duplicate") {
			err = fmt.Errorf("Game %s already created", gameid)
		} else {
			err = fmt.Errorf("Mongodb issue on GameCreated event write: %s", mongoerr.Error())
		}
		return
	}

	// Create and register the game object in the game pool
	game := types.NewGameFromEvent(ev)
	if gperr := h.gPool.AddGame(game); gperr != nil {
		// Should catch all dups at the event level
		if strings.Contains(gperr.Error(), "duplicate") {
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
		err = fmt.Errorf("The requested GameID: %s is not accepting players. State=%s", gameid, game.Status)
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
			err = fmt.Errorf("Something bad happened. PlayerPool out of sync with mongo events")
			return
		}
		err = fmt.Errorf("Issue on AddPlayer add to PlayerPool: %s", hperr.Error())
		return
	}
	// Persist the pPool, or should it auto persist on state change?

	// I guess it makes sense to increment the game player count at this point.
	game.StartPlayers++
	return nil
}

// OnGameStarted handles activiting a game from the starting stage into playing.
// Only the original game creator is allowed to start a given gameid.
// Errors:
// -- gameid or slackid empty
// -- gameid not exists and in 'starting' state
// -- slackid does not match the creating slackid
func (h *Handler) OnGameStarted(gameid string, slackid string) (err error) {
	// First, make sure there's already a game and it's not started yet
	err = h.gPool.StartGame(gameid, slackid)
	return
}

// GetGameStatus produces a game status report for the specified gameid
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
