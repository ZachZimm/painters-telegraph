// Create a web server that listens on port 8080 and returns a string "Hello world"
// when the endpoint "/" is hit.
// Use the net/http package to create the server.
// Use the ListenAndServe method to listen on port 8080.
package main

import (
	"fmt"
	"net/http"
)

// Things I need for a game server
// A map of ongoing games - even though I only intend to have one game at a time for now
// A game struct that contains the game state
// A player struct that contains the player state

var games map[string]Game = make(map[string]Game)

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
		currentRound: 1,
		promptsSet:   false,
		players:      []Player{},
		prompts:      [][]string{},
		drawings:     [][]string{},
	}

	// Add the game to the games map
	games[game.gameName] = game

	fmt.Fprintf(w, "Game created")
}

func listGames(w http.ResponseWriter, r *http.Request) {
	// Loop through the games map and return the game names
	for key, _ := range games {
		fmt.Fprintf(w, key)
	}
}

func gameStateToJSON(game Game) string {
	// Convert the game state to JSON
	// gameJSON, err := json.Marshal(game)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	gameJsonString := ""
	gameJsonString += "{"
	gameJsonString += "\"gameName\": \"" + game.gameName + "\","
	gameJsonString += "\"totalRounds\": " + fmt.Sprint(game.totalRounds) + ","
	gameJsonString += "\"currentRound\": " + fmt.Sprint(game.currentRound) + ","
	gameJsonString += "\"promptsSet\": " + fmt.Sprint(game.promptsSet) + ","
	gameJsonString += "\"players\": ["
	for i, player := range game.players {
		gameJsonString += "{"
		gameJsonString += "\"playerName\": \"" + player.playerName + "\","
		gameJsonString += "\"playerID\": " + fmt.Sprint(player.playerID) + ","
		gameJsonString += "\"playerSecret\": \"" + player.playerSecret + "\""
		gameJsonString += "}"
		if i < len(game.players)-1 {
			gameJsonString += ","
		}
	}
	gameJsonString += "],"
	gameJsonString += "\"prompts\": ["
	for i, prompt := range game.prompts {
		gameJsonString += "["
		for j, _ := range prompt {
			gameJsonString += "\"" + game.prompts[i][j] + "\""
			if j < len(prompt)-1 {
				gameJsonString += ","
			}
		}
	}
	gameJsonString += "],"
	gameJsonString += "\"drawings\": ["
	for i, drawing := range game.drawings {
		gameJsonString += "["
		for j, _ := range drawing {
			gameJsonString += "\"" + game.drawings[i][j] + "\""
			if j < len(drawing)-1 {
				gameJsonString += ","
			}
		}
	}
	gameJsonString += "]"
	gameJsonString += "}"
	return gameJsonString
}

func getGameState(w http.ResponseWriter, r *http.Request) {
	// Example URL: http://localhost:9119/getGameState?gameName=test
	gameName := r.URL.Query().Get("gameName")
	game := games[gameName]
	gameStateJSON := gameStateToJSON(game)
	fmt.Fprintf(w, gameStateJSON)
}

func endGame(w http.ResponseWriter, r *http.Request) {
	// TODO Implement logic for ending a game and creating GIFs, as well as a final game state that can be reviewed
	gameName := r.URL.Query().Get("gameName")
	fmt.Fprintf(w, "Game ended")
	// Remove the game from the games map
	delete(games, gameName)
}

func endRound(w http.ResponseWriter, r *http.Request) {
	// End a round
	gameName := r.URL.Query().Get("gameName")
	game := games[gameName]
	if game.promptsSet == false {
		if game.currentRound == game.totalRounds {
			endGame(w, r)
		} else {
			game.promptsSet = true
		}
	} else {
		if game.currentRound != game.totalRounds {
			game.currentRound++
			game.promptsSet = false
		}
	}

	// Implement logic for either serving prompts or drawings to players
	// This will be done using the round number as an offset to the player index
	// First round player 1 creates prompt 1, player 2 gets prompt 1, player 3 gets prompt 2, etc.
	fmt.Fprintf(w, "Round ended")

	// Remove the game from the games map if the game is over
}

// An endpoint to end a game
// An endpoint to end a round
// An endpoint to get a game's current game state
// - An endpoint to get all game names
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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})

	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("/listGames", listGames)
	http.HandleFunc("/getGameState", getGameState)
	http.HandleFunc("/endGame", endGame)
	http.HandleFunc("/endRound", endRound)

	http.ListenAndServe(":9119", nil)
}
