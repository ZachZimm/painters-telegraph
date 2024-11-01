package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"

	"golang.org/x/image/draw"
)

var games map[string]*Game = make(map[string]*Game)
var players map[string]*Player = make(map[string]*Player)

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

func getPlayerIndex(playerName string, game *Game) int {
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
	game := &Game{
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
	gameStateJSON := gameStateToJSON(*game)
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

		fmt.Fprintf(w, "Game started")
	} else {
		fmt.Fprintf(w, "Game already started")
	}
}

func _endGame(gameName string) {
	// TODO Implement logic for ending a game and creating GIFs, as well as a final game state that can be reviewed
	delete(games, gameName)
}

func endGame(w http.ResponseWriter, r *http.Request) {
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	fmt.Fprintf(w, "Game ended")
	_endGame(gameName)
}

func _endRound(gameName string) bool {
	game, ok := games[gameName]
	if !ok {
		return false
	}

	if game.promptsSet == false {
		if game.currentRound == game.totalRounds {
			_endGame(gameName)
		} else {
			game.promptsSet = true
			return true
		}
	} else {
		if game.currentRound != game.totalRounds {
			game.currentRound++
			game.promptsSet = false
			return true
		}
	}
	return false
}

func endRound(w http.ResponseWriter, r *http.Request) {
	// End a round
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	_endRound(gameName)
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
		players[givenPlayerName] = &Player{playerName: givenPlayerName, playerSecret: givenPlayerSecret}
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
			game.players = append(game.players, *player)
			games[gameName] = game // I don't like this approach
			fmt.Fprintf(w, "Player joined game %s", gameName)
		} else {
			game.spectators = append(game.spectators, *player)
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
		} else {
			fmt.Fprintf(w, "Drawing already submitted")
		}
	} else {
		fmt.Fprintf(w, "Player not authenticated")
		return
	}
}

func generateShortHash() string {
	b := make([]byte, 8) // 8 bytes = 64 bits
	_, err := rand.Read(b)
	if err != nil {
		// Handle error appropriately
		return ""
	}
	return hex.EncodeToString(b)
}

func resizeImage(img image.Image, width, height int) image.Image {
	// Create a new RGBA image with the desired dimensions
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Use the Nearest-Neighbor algorithm for scaling
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	// For better quality, you can use Catmull-Rom interpolation:
	// draw.CatmullRom.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	return dst
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	} else if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}

	host := r.Host

	return fmt.Sprintf("%s://%s", scheme, host)
}

func uploadDrawing(w http.ResponseWriter, r *http.Request) {
	// Limit the size of the uploaded file to 10 MB
	err := r.ParseMultipartForm(10 << 20) // We are limiting the size to 10 MB
	if err != nil {
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	// Get form values for authentication
	playerName := r.FormValue("playerName")
	playerSecret := r.FormValue("playerSecret")

	// Authenticate the player
	if !authenticatePlayer(playerName, playerSecret) {
		http.Error(w, "Player not authenticated", http.StatusUnauthorized)
		return
	}

	// Retrieve the file from the request
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Error decoding image", http.StatusBadRequest)
		return
	}
	if format != "jpeg" && format != "png" {
		http.Error(w, "Unsupported file type", http.StatusUnsupportedMediaType)
		return
	}

	resizedImage := resizeImage(img, 1024, 1024)

	// Generate a short hash for the filename
	hash := generateShortHash()
	if hash == "" {
		http.Error(w, "Error generating file name", http.StatusInternalServerError)
		return
	}

	// Ensure the images directory exists
	if _, err := os.Stat("images"); os.IsNotExist(err) {
		err = os.Mkdir("images", os.ModePerm)
		if err != nil {
			http.Error(w, "Error creating images directory", http.StatusInternalServerError)
			return
		}
	}

	// Create the output file
	outputPath := fmt.Sprintf("images/%s.png", hash)
	outFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "Unable to create the file for writing", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	// Save the image in PNG format
	err = png.Encode(outFile, resizedImage)
	if err != nil {
		http.Error(w, "Error encoding image to PNG", http.StatusInternalServerError)
		return
	}

	// Return the relative URL of the saved image
	baseUrl := getBaseURL(r)
	imageUrl := fmt.Sprintf("%s/%s", baseUrl, outputPath)
	fmt.Fprintf(w, imageUrl)
}

// -An endpoint to end a game
// -An endpoint to end a round
// -An endpoint to get a game's current game state
// -An endpoint to get all game names
// -An endpoint to authenticate a player
// -An endpoint for an authenticated player to join a game
// -An endpoint for an authenticated player to submit an intial prompt
// -An endpoint for an authenticated player to submit a drawing
// -An endpoint for an authenticated player to submit a caption
// An endpoint for uploading images to the server which returns a URL for the image
//
// Functions for queueing content to display to players when requested
// 		A map of player names to queued messages in JSON format. If the message is an image, the frontend will handle it
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
	http.HandleFunc("/uploadDrawing", uploadDrawing)

	fs := http.FileServer(http.Dir("images"))
	http.Handle("/images/", http.StripPrefix("/images/", fs))
	// example: http://localhost:9119/images/12345678.png

	http.ListenAndServe(":9119", nil)
}
