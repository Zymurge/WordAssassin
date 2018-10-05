package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	dao "wordassassin/persistence"
	types "wordassassin/types"
)

const (
	mongoDB           string = "wordDB"
	defaultPort       string = "8080"
	serverPortEnvName string = "PORT"
	mongoURLEnvName   string = "MONGOURL"
)

var (
	port     string
	mongoURL string
	logger   *log.Logger
	mongo    *dao.MongoSession
	games    *types.GamePool
	players  types.PlayerPool
	handler  Handler
)

func getGamesList(c echo.Context) error {
	return c.HTML(http.StatusOK, handler.GetGamesList())
}

func getGameStatus(c echo.Context) error {
	gameid := c.Param("gameid")
	if message, exists := handler.GetGameStatus(gameid); exists {
		return c.HTML(http.StatusOK, message)
	}
	message := fmt.Sprintf("Game %s not found", gameid)
	return c.HTML(http.StatusNotFound, message)
}

func createGame(c echo.Context) error {
	gameid := c.Param("gameid")
	creator := c.Param("creator")
	var (
		killdict string // get from query
		passcode string // get from query
	)

	if err := handler.OnGameCreated(gameid, creator, killdict, passcode); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	message := fmt.Sprintf("<h3>Game Created</h3><p>Game: %s  Creator: %s", gameid, creator)
	return c.HTML(http.StatusOK, message)
}

func addPlayer(c echo.Context) error {
	// TODO: pass all vars (tag, name, email)
	gameid := c.Param("gameid")
	slackid := c.Param("slackid")
	name := c.Param("name")
	email := c.Param("email")
	if err := handler.OnPlayerAdded(gameid, slackid, name, email); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	message := fmt.Sprintf("Player %s added to game %s", slackid, gameid)
	return c.HTML(http.StatusOK, message)
}

func healthCheck(c echo.Context) error {
	return c.HTML(http.StatusOK, "I'm running!")
}

func main() {
	var err error
	logger = log.New(os.Stderr, "WordAssassin: ", log.Ldate|log.Ltime)
	// Port from env or default
	if port = os.Getenv(serverPortEnvName); port == "" {
		port = defaultPort
	}
	port = ":" + port

	// MongoURL or bail if not available
	if mongoURL = os.Getenv(mongoURLEnvName); mongoURL == "" {
		logger.Fatalf("Can't find Mongo URL env variable: %s", mongoURLEnvName)
	}
	mongo, err = dao.NewMongoSession(mongoURL, mongoDB, logger)
	if err != nil { log.Panicf("NewMongoSession: %s", err)}

	players = types.PlayerPool{}
	games = types.NewGamePool(mongo)
	handler = NewHandler(games, &players, mongo)

	//*** Web Server Stuff ***//
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", healthCheck)
	e.GET("/health", healthCheck)
	e.GET("/gamestatus/:gameid", getGameStatus)
	e.GET("/gameslist", getGamesList)
	e.POST("/creategame/:gameid/:creator", createGame)
	e.POST("/addplayer/:gameid/:slackid", addPlayer)

	// Start server
	e.Logger.Fatal(e.Start(port))
}
