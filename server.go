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
	mongoURL              string = "localhost:27017"
	mongoDB               string = "testDB"
	mongoPlayerCollection string = "players"
	serverPortEnvName	  string = "PORT"
	defaultPort			  string = "8080"
)

var (
	port	string
	logger  *log.Logger
	mongo   *dao.MongoSession
	players types.PlayerPool
	handler Handler
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
	if port = os.Getenv(serverPortEnvName); port == "" {
		port = defaultPort
	}
	port = ":" + port
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
