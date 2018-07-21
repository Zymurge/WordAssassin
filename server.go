package main

import (
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
	players  types.PlayerPool
	handler  Handler
)

func addPlayer(c echo.Context) error {
	// TODO: pass all vars (tag, name, email)
	player := c.Param("tag")
	if err := handler.OnPlayerAdded(player, player, "dummy@email.org"); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	message := "Player added: " + player
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

	// TODO: rethink the gameID mapping per pool, but hardcode one for now
	players = types.PlayerPool{GameID: "testGameID"}
	handler = NewHandler(&players, mongo)

	//*** Web Server Stuff ***//
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", healthCheck)
	e.POST("/addplayer/:tag", addPlayer)

	// Start server
	e.Logger.Fatal(e.Start(port))
}
