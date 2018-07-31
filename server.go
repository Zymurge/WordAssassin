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
	games    types.GamePool
	players  types.PlayerPool
	handler  Handler
)

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
	return c.HTML(400, message)
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
	return c.HTML(200, message)
}

func healthCheck(c echo.Context) error {
	return c.HTML(http.StatusOK, "I'm running!")
}

func main() {
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
	mongo = dao.NewMongoSession(mongoURL, mongoDB, logger)

	players = types.PlayerPool{}
	games = types.GamePool{}
	handler = NewHandler(&games, &players, mongo)

	//*** Web Server Stuff ***//
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", healthCheck)
	e.GET("/health", healthCheck)
	e.POST("/creategame/:gameid/:creator", createGame)
	e.POST("/addplayer/:gameid/:slackid", addPlayer)

	// Start server
	e.Logger.Fatal(e.Start(port))
}
