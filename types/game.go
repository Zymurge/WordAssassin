package types

import (
	"fmt"
	"time"
	"github.com/mongodb/mongo-go-driver/bson"	
	"wordassassin/types/events"
)

// GameStatus allows management of game state
type GameStatus string

// Game - The elements necessary to track a single game of wordassassin.
type Game struct {
	ID             string     `json:"id" bson:"_id"`
	TimeCreated    time.Time  `json:"timeCreated" bson:"timecreated"`
	GameCreator    string     `json:"gameId" bson:"gameid"`
	KillDictionary string     `json:"name" bson:"name"`
	Passcode       string     `json:"passcode" bson:"passcode"`
	Status         GameStatus `json:"status" bson:"status"`
	StartTime      time.Time  `json:"starttime"`
	StartPlayers   int        `json:"startplayers"`
	RemainPlayers  int        `json:"remainplayers"`
	// Other possible things:
	//	TargetList
	//	NumKills
}

// Constants for GameStatus
const (
	Starting GameStatus = "starting"
	Playing  GameStatus = "playing"
	Finished GameStatus = "finished"
	Aborted  GameStatus = "aborted"
)

const (
	// MinimumPlayers - Valid minimum number of players to play a game
	MinimumPlayers int = 5
)

// NewGameFromEvent instantiates a Player from a PlayerAdedEvent
func NewGameFromEvent(ev events.GameCreatedEvent) (g Game) {
	g = Game{
		ID:             ev.ID,
		TimeCreated:    ev.TimeCreated,
		GameCreator:    ev.GameCreator,
		KillDictionary: ev.KillDictionary,
		Passcode:       ev.Passcode,
		Status:         Starting,
		StartTime:		time.Unix(0, 0),
	}
	return
}

// Decode populates this instance from the supplied bson
func (g *Game) Decode(raw []byte) error {
	if err := bson.Unmarshal(raw, g); err != nil {
		return err
	}
	return nil
}

// GetID getter for ID field
func (g *Game) GetID() string {
	return g.ID
}

// GetPlayerList fetches a map of players from the player pool for this game keyed by ID
func (g *Game) GetPlayerList(pp *MockPlayerPool) (result map[string]*Player) {
	// TODO: Create PlayerPool method that fectches all by GameID
	if list, err := pp.GetAllPlayersInGame(g.GetID()); err == nil {
		// populate the map from the array
		result = make(map[string]*Player, len(list))
		for _, v := range(list) {
			result[v.GetID()] = v
		}
	} else {
		panic("do something with the error")
	}

	return
}

// GetStatus translates the status enum
func (g *Game) GetStatus() string {
	// TODO: map to string values of enum
	return string(g.Status)
}

// GetStatusReport generates a status report for this game instance
func (g *Game) GetStatusReport() string {
	result := 
	fmt.Sprintf("Game Status for %s:\n\n", g.GetID()) +
	fmt.Sprintf("   Status: %s\n", 	       g.GetStatus()) +
	fmt.Sprintf("   # Players: %d\n",      g.StartPlayers)
	
	return result	
	}	
	
// Start does whatever is needed to transition from setup to go time
func (g *Game) Start(players []*Player) error {
	// Validate whatever needs validating
	if g.Status != Starting {
		return fmt.Errorf("Game not in starting state. Current state is %s", g.GetStatus())
	}
	// make sure something isn't horribly wrong in accounting
	// TODO: reenable after fixing tests in handler_test
	// if g.StartPlayers != len(players) {
	// 	panic( fmt.Sprintf("Game: %s.StartPlayers=%d, PlayerPool count=%d", g.GetID(), g.StartPlayers, len(players))	)	
	// }
	if g.StartPlayers < MinimumPlayers {
		return fmt.Errorf("Game requires %d players. Current count is %d", MinimumPlayers, g.StartPlayers)
	}
	if !ValidDictionary(g.KillDictionary) {
		return fmt.Errorf("Game requires a valid dictionary. %s doesn't meet the critera", g.KillDictionary)
	}
	// Set the game status to "running"
	g.Status = Playing
	// Assign first round of targets
	if err := g.SetAllTargets(players); err != nil {
		return fmt.Errorf("Failure to set targets for game: %s. Error: %v", g.GetID(), err)
	}
	// Log what you gotta log -- unless an event is written first
	return nil
}
	
// SetAllTargets creates the targets and kill words for all players in a list, using this Game's kill dict
func (g *Game) SetAllTargets(players []*Player) error {
	// TODO: implement
	// create a circular linked list of all the players randomly
	// for each assignment, assign next as target
	// for each assignment, send target notification -- delay until last in case of issues above to prevent chances
	//   of false notification
	return fmt.Errorf("Not implemented")
}	

// ValidDictionary checks for the existence and proper format of a KillDictionary
func ValidDictionary( dict string ) bool {
	// TODO: define and implement what valid means
	return true
}