package main

import (
	"net/http"
	"log"
	"os"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	types "wordassassin/types"
	dao "wordassassin/persistence"
)

const (
	mongoURL				string = "localhost:27017"
	mongoDB    				string = "testDB"
	mongoPlayerCollection	string = "players"
)

var (
	logger *log.Logger
	mongo *dao.MongoSession
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

func main() {
	logger = log.New(os.Stderr, "WordAssassin: ", log.Ldate|log.Ltime)
	mongo = dao.NewMongoSession( mongoURL, mongoDB, logger	)
	// TODO: rethink the gameID mapping per pool, but hardcode one for now
	players = types.PlayerPool{ GameID: "testGameID" }
	handler = NewHandler(&players, mongo)

	//*** Web Server Stuff ***//
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/addplayer/:tag", addPlayer)

	// Start server
	e.Logger.Fatal(e.Start(":1313"))
}
