// Create a web server that listens on port 8080 and returns a string "Hello world"
// when the endpoint "/" is hit.
// Use the net/http package to create the server.
// Use the ListenAndServe method to listen on port 8080.
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})

	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("listGames", listGames)

	http.ListenAndServe(":9119", nil)
}

// Things I need for a game server
// A map of ongoing games - even though I only intend to have one game at a time for now
// A game struct that contains the game state
// A player struct that contains the player state

var games map[string]Game

type Player struct {
	playerName   string
	playerID     int // might not need this
	playerSecret string
}

// Game struct
type Game struct {
	gameName     string
	totalRounds  int
	currentRound int
	promptsSet   bool
	players      []Player
	prompts      [][]string
	drawings     [][]string
}

// An endpoint to create a new game
func createGame(w http.ResponseWriter, r *http.Request) {
	// Create a new game
	var game Game = Game{
		gameName:     "test",
		totalRounds:  5,
		currentRound: 0,
		promptsSet:   false,
		players:      []Player{},
		prompts:      [][]string{},
		drawings:     [][]string{},
	}

	// Add the game to the games map
	games[game.gameName] = game
}

func listGames(w http.ResponseWriter, r *http.Request) {
	// Loop through the games map and return the game names
	for key, _ := range games {
		fmt.Fprintf(w, key)
	}
}

// An endpoint to end a game
// An endpoint to end a round
// An endpoint to get a game's current game state
// An endpoint to get all game names
// An endpoint to authenticate a player
// An endpoint for an authenticated player to join a game
// An endpoint for an authenticated player to submit an intial prompt
// An endpoint for an authenticated player to submit a drawing
// An endpoint for an authenticated player to submit a caption
//
// Functions for when the game starts
// Functions for passing prompts around
// Functions for passing drawings around
// Functions for when the game ends
// Function for resizing uploaded images
// Function for creating GIFs from images and captions
