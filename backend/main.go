package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var games map[string]Game = make(map[string]Game)
var players map[string]Player = make(map[string]Player)

type Player struct {
	playerName   string
	playerSecret string
}

// Game struct
type Game struct {
	gameName     string
	totalRounds  int
	currentRound int
	promptsSet   bool
	gameStarted  bool
	players      []Player
	spectators   []Player
	prompts      [][]string
	drawings     [][]string
}

func getPlayerIndex(playerName string, game Game) int {
	for i, player := range game.players {
		if player.playerName == playerName {
			return i
		}
	}
	return -1
}

func parseBodyObject(r *http.Request) map[string]string {
	body := make([]byte, r.ContentLength)
	r.Body.Read(body)
	bodyObj := make(map[string]string)
	json.Unmarshal(body, &bodyObj)
	return bodyObj
}

// An endpoint to create a new game
func createGame(w http.ResponseWriter, r *http.Request) {
	// Create a new game
	jsonObject := parseBodyObject(r)
	var game Game = Game{
		gameName:     jsonObject["gameName"],
		totalRounds:  5,
		currentRound: 0,
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
	gameJsonString := ""
	gameJsonString += "{"
	gameJsonString += "\"gameName\": \"" + game.gameName + "\","
	gameJsonString += "\"totalRounds\": " + fmt.Sprint(game.totalRounds) + ","
	gameJsonString += "\"currentRound\": " + fmt.Sprint(game.currentRound) + ","
	gameJsonString += "\"promptsSet\": " + fmt.Sprint(game.promptsSet) + ","
	gameJsonString += "\"gameStarted\": " + fmt.Sprint(game.gameStarted) + ","
	gameJsonString += "\"players\": ["
	for i, player := range game.players {
		gameJsonString += "{"
		gameJsonString += "\"playerName\": \"" + player.playerName + "\""
		gameJsonString += "}"
		if i < len(game.players)-1 {
			gameJsonString += ","
		}
	}
	gameJsonString += "],"
	gameJsonString += "\"spectators\": ["
	for i, spectator := range game.spectators {
		gameJsonString += "{"
		gameJsonString += "\"playerName\": \"" + spectator.playerName + "\""
		gameJsonString += "}"
		if i < len(game.spectators)-1 {
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
		gameJsonString += "]"
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
		gameJsonString += "]"
	}
	gameJsonString += "]"
	gameJsonString += "}"
	return gameJsonString
}

func getGameState(w http.ResponseWriter, r *http.Request) {
	jsonObject := parseBodyObject(r)
	game, ok := games[jsonObject["gameName"]]
	if !ok {
		fmt.Fprintf(w, "Game not found")
		return
	}
	gameStateJSON := gameStateToJSON(game)
	fmt.Fprintf(w, gameStateJSON)
}

func startGame(w http.ResponseWriter, r *http.Request) {
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	game, ok := games[gameName]
	if !ok {
		fmt.Fprintf(w, "Game not found")
		return
	}

	if game.gameStarted == false {
		game.gameStarted = true
		game.prompts = make([][]string, len(game.players))
		game.drawings = make([][]string, len(game.players))
		for i := range game.prompts {
			game.prompts[i] = make([]string, game.totalRounds)
			game.drawings[i] = make([]string, game.totalRounds)
		}

		games[gameName] = game
		fmt.Fprintf(w, "Game started")
	} else {
		fmt.Fprintf(w, "Game already started")
	}
}

func endGame(w http.ResponseWriter, r *http.Request) {
	// TODO Implement logic for ending a game and creating GIFs, as well as a final game state that can be reviewed
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	fmt.Fprintf(w, "Game ended")
	// Remove the game from the games map
	delete(games, gameName)
}

func endRound(w http.ResponseWriter, r *http.Request) {
	// End a round
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	game, ok := games[gameName]
	if !ok {
		fmt.Fprintf(w, "Game not found")
		return
	}

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

	games[gameName] = game

	// Implement logic for either serving prompts or drawings to players
	// This will be done using the round number as an offset to the player index
	// First round player 1 creates prompt 1, player 2 gets prompt 1, player 3 gets prompt 2, etc.
	fmt.Fprintf(w, "Round ended")

	// Remove the game from the games map if the game is over
}

func authenticatePlayer(givenPlayerName, givenPlayerSecret string) bool {
	if givenPlayerName == "" {
		return false
	}
	existingPlayer, playerExists := players[givenPlayerName]
	if playerExists {
		if existingPlayer.playerSecret == givenPlayerSecret {
			return true

		} else {
			return false
		}
	} else {
		players[givenPlayerName] = Player{playerName: givenPlayerName, playerSecret: givenPlayerSecret}
		return true
	}
}

func checkAuthentication(w http.ResponseWriter, r *http.Request) {
	// Check if a player is authenticated
	jsonObject := parseBodyObject(r)
	playerName := jsonObject["playerName"]
	playerSecret := jsonObject["playerSecret"]
	returnString := "{"
	if authenticatePlayer(playerName, playerSecret) {
		returnString += "\"authenticated\": true}"
	} else {
		returnString += "\"authenticated\": false}"
	}
	fmt.Fprintf(w, returnString)
}

func joinGame(w http.ResponseWriter, r *http.Request) {
	// Accept a POST request to join a game
	bodyObj := parseBodyObject(r)

	gameName := bodyObj["gameName"]
	playerName := bodyObj["playerName"]
	playerSecret := bodyObj["playerSecret"]

	if authenticatePlayer(playerName, playerSecret) {
		game, ok := games[gameName]
		if !ok {
			fmt.Fprintf(w, "Game not found")
			return
		}
		player := players[playerName]
		if game.gameStarted == false {
			game.players = append(game.players, player)
			games[gameName] = game // I don't like this approach
			fmt.Fprintf(w, "Player joined game %s", gameName)
		} else {
			game.spectators = append(game.spectators, player)
			fmt.Fprintf(w, "Player joined game as spectator")
		}
	} else {
		fmt.Fprintf(w, "Player not authenticated")
		return
	}
}

func submitPrompt(w http.ResponseWriter, r *http.Request) {
	// Submit a prompt to the current game
	bodyObj := parseBodyObject(r)

	gameName := bodyObj["gameName"]
	playerName := bodyObj["playerName"]
	playerSecret := bodyObj["playerSecret"]
	prompt := bodyObj["prompt"]

	if authenticatePlayer(playerName, playerSecret) {
		game, ok := games[gameName]
		if !ok {
			fmt.Fprintf(w, "Game not found")
			return
		}

		if game.gameStarted == false {
			fmt.Fprintf(w, "Game not started")
			return
		}
		if game.promptsSet == true {
			fmt.Fprintf(w, "Prompts already set for this round")
			return
		}
		playerIndex := getPlayerIndex(playerName, game)
		if playerIndex == -1 {
			fmt.Fprintf(w, "Player not in game")
			return
		}

		gameRotationIndex := (playerIndex + game.currentRound) % len(game.players)
		if len(game.prompts) == 0 || len(game.prompts[gameRotationIndex]) == 0 {
			fmt.Fprintf(w, "Game prompts slice not initialized")
			return
		}
		if game.prompts[gameRotationIndex][game.currentRound] == "" {
			game.prompts[gameRotationIndex][game.currentRound] = prompt
			fmt.Fprintf(w, "Prompt submitted")
			games[gameName] = game
		} else {
			fmt.Fprintf(w, "Prompt already submitted")
		}
	} else {
		fmt.Fprintf(w, "Player not authenticated")
		return
	}
}

func submitDrawing(w http.ResponseWriter, r *http.Request) {
	// Submit a prompt to the current game
	bodyObj := parseBodyObject(r)

	gameName := bodyObj["gameName"]
	playerName := bodyObj["playerName"]
	playerSecret := bodyObj["playerSecret"]
	drawing := bodyObj["drawing"]

	if authenticatePlayer(playerName, playerSecret) {
		game, ok := games[gameName]
		if !ok {
			fmt.Fprintf(w, "Game not found")
			return
		}

		if game.gameStarted == false {
			fmt.Fprintf(w, "Game not started")
			return
		}
		if game.promptsSet == false {
			fmt.Fprintf(w, "Prompts not yet set for this round")
			return
		}
		playerIndex := getPlayerIndex(playerName, game)
		if playerIndex == -1 {
			fmt.Fprintf(w, "Player not in game")
			return
		}

		gameRotationIndex := (playerIndex + game.currentRound) % len(game.players)
		if len(game.drawings) == 0 || len(game.drawings[gameRotationIndex]) == 0 {
			fmt.Fprintf(w, "Game drawings slice not initialized")
			return
		}
		if game.drawings[gameRotationIndex][game.currentRound] == "" {
			game.drawings[gameRotationIndex][game.currentRound] = drawing
			fmt.Fprintf(w, "Drawing submitted")
			games[gameName] = game
		} else {
			fmt.Fprintf(w, "Drawing already submitted")
		}
	} else {
		fmt.Fprintf(w, "Player not authenticated")
		return
	}
}

// -An endpoint to end a game
// -An endpoint to end a round
// -An endpoint to get a game's current game state
// -An endpoint to get all game names
// -An endpoint to authenticate a player
// -An endpoint for an authenticated player to join a game
// -An endpoint for an authenticated player to submit an intial prompt
// An endpoint for an authenticated player to submit a drawing
// -An endpoint for an authenticated player to submit a caption
//
// Functions for queueing content to display to players when requested
// Functions for when the game starts
// Routine for automatically progressing the game based on a configurable timer
// Functions for passing prompts around
// Functions for passing drawings around
// Functions for when the game ends
// Function for resizing uploaded images
// Function for creating GIFs from images and captions

func main() {
	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("/listGames", listGames)
	http.HandleFunc("/getGameState", getGameState)
	http.HandleFunc("/startGame", startGame)
	http.HandleFunc("/endGame", endGame)
	http.HandleFunc("/endRound", endRound)
	http.HandleFunc("/checkAuthentication", checkAuthentication)
	http.HandleFunc("/joinGame", joinGame)
	http.HandleFunc("/submitPrompt", submitPrompt)
	http.HandleFunc("/submitDrawing", submitDrawing)
	http.ListenAndServe(":9119", nil)
}
