package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
)

var games map[string]*Game = make(map[string]*Game)
var endedGames map[string]*EndedGame = make(map[string]*EndedGame)
var players map[string]*Player = make(map[string]*Player)
var (
	newPlayerMessage               = "{\"status\": \"OK\", \"message\":\"You have not yet joined a game\"}"
	joinedGameMessage              = "{\"status\": \"OK\", \"message\":\"You have joined the game\"}"
	gameStartedMessage             = "{\"status\": \"OK\", \"message\":\"Write an interesting prompt!\"}"
	drawPromptMessage              = "{\"status\": \"OK\", \"message\":\"Draw the prompt!\""
	captionPromptMessage           = "{\"status\": \"OK\", \"message\":\"Write a caption for the drawing!\""
	gameEndedMessage               = "{\"status\": \"OK\", \"message\":\"The game has ended, check the results!\""
	nonSubmissionString_drawing    = "Uh oh. Looks like someone forgot to submit their drawing =/"
	nonSubmissionString_caption    = "Uh oh. Looks like someone forgot to submit their caption =/"
	fontName                       = "Roboto-Regular.ttf"
	imagesDir                      = "images"
	nonSubmissionImageName_drawing = "non_submission_drawing.png"
	nonSubmissionImageName_caption = "non_submission_caption.png"
	baseUrl                        = ""
)

type Player struct {
	playerName    string
	playerSecret  string
	queuedMessage string
}

// Game struct
type Game struct {
	gameName     string
	gameId       string
	roundTimer   int
	totalRounds  int
	currentRound int
	promptsSet   bool
	gameStarted  bool
	players      []*Player
	spectators   []*Player
	prompts      [][]string
	drawings     [][]string
}

type EndedGame struct {
	gameName string
	gameId   string
	prompts  [][]string
	drawings [][]string
	gifs     []string
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

func getPlayerQueuedMessage(w http.ResponseWriter, r *http.Request) {
	jsonObject := parseBodyObject(r)
	playerName := jsonObject["playerName"]
	player, err := players[playerName]
	if !err {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Player not found\"}"
		fmt.Fprintf(w, responseStr)
		return
	}
	fmt.Fprintf(w, player.queuedMessage)
}

// An endpoint to create a new game
func createGame(w http.ResponseWriter, r *http.Request) {
	if baseUrl == "" {
		baseUrl = getBaseURL(r)
	}

	// Create a new game
	jsonObject := parseBodyObject(r)
	// parse the JSON object for the fields, and use default values if they are not provided
	_gameName := jsonObject["gameName"]
	if _gameName == "" {
		fmt.Fprintf(w, "{\"status\": \"ERROR\", \"message\": \"gameName is required\"}")
		return
	}
	// Check if the game already exists
	// If the game already exists, return an error
	if _, ok := games[_gameName]; ok {
		fmt.Fprintf(w, "{\"status\": \"ERROR\", \"message\": \"Game "+jsonObject["gameName"]+" already exists\"}")
		return
	}
	if jsonObject["totalRounds"] == "" {
		jsonObject["totalRounds"] = "3"
	}
	// try to parse an int from the totalRounds field
	_totalRounds, err := strconv.Atoi(jsonObject["totalRounds"])
	if err != nil {
		fmt.Fprintf(w, "{\"status\": \"ERROR\", \"message\": \"totalRounds must be an integer\"}")
		return
	}
	if jsonObject["roundTimer"] == "" {
		jsonObject["roundTimer"] = "60"
	}
	_roundTimer, err := strconv.Atoi(jsonObject["roundTimer"])
	if err != nil {
		fmt.Fprintf(w, "{\"status\": \"ERROR\", \"message\": \"roundTimer must be an integer\"}")
		return
	}

	hash := make([]byte, 16)
	rand.Read(hash)
	_gameId := hex.EncodeToString(hash)

	game := &Game{
		gameName:     _gameName,
		gameId:       _gameId,
		roundTimer:   _roundTimer,
		totalRounds:  _totalRounds,
		currentRound: 0,
		promptsSet:   false,
		players:      []*Player{},
		spectators:   []*Player{},
		prompts:      [][]string{},
		drawings:     [][]string{},
	}

	// Add the game to the games map
	games[game.gameName] = game

	fmt.Fprintf(w, "{\"status\": \"OK\", \"message\": \"Game "+game.gameName+" created\"}")
}

func listGames(w http.ResponseWriter, r *http.Request) {
	// Loop through the games map and return the game names
	responseStr := "{\"status\":\"OK\", \"games\": ["
	for key := range games {
		responseStr += "\"" + key + "\","
	}
	if len(games) > 0 {
		responseStr = responseStr[:len(responseStr)-1]
	}
	responseStr += "]}"
	fmt.Fprintf(w, responseStr)
}

func listEndedGames(w http.ResponseWriter, r *http.Request) {
	// Loop through the endedGames map and return the game ids
	responseStr := "{\"status\":\"OK\", \"games\": ["
	for key := range endedGames {
		responseStr += "\"" + key + "\","
	}
	if len(endedGames) > 0 {
		responseStr = responseStr[:len(responseStr)-1]
	}
	responseStr += "]}"
	fmt.Fprintf(w, responseStr)
}

func endedGameStateToJSON(endedGame EndedGame) string {
	gameJsonString := ""
	gameJsonString += "{"
	gameJsonString += "\"gameName\": \"" + endedGame.gameName + "\","
	gameJsonString += "\"gameId\": \"" + endedGame.gameId + "\","
	gameJsonString += "\"prompts\": ["
	for _, prompt := range endedGame.prompts {
		gameJsonString += "["
		for _, word := range prompt {
			gameJsonString += "\"" + word + "\","
		}
		gameJsonString = gameJsonString[:len(gameJsonString)-1]
		gameJsonString += "],"
	}
	if len(endedGame.prompts) > 0 {
		gameJsonString = gameJsonString[:len(gameJsonString)-1]
	}
	gameJsonString += "],"
	gameJsonString += "\"drawings\": ["
	for _, drawing := range endedGame.drawings {
		gameJsonString += "["
		for _, word := range drawing {
			gameJsonString += "\"" + word + "\","
		}
		gameJsonString = gameJsonString[:len(gameJsonString)-1]
		gameJsonString += "],"
	}
	if len(endedGame.drawings) > 0 {
		gameJsonString = gameJsonString[:len(gameJsonString)-1]
	}
	gameJsonString += "],"
	gameJsonString += "\"gifs\": ["
	for _, gif := range endedGame.gifs {
		gameJsonString += "["
		gameJsonString += "\"" + gif + "\","
		gameJsonString += "],"
	}
	if len(endedGame.gifs) > 0 {
		gameJsonString = gameJsonString[:len(gameJsonString)-1]
	}
	gameJsonString += "]"
	gameJsonString += "}"
	return gameJsonString
}

func gameStateToJSON(game Game) string {
	// Convert the game state to JSON
	gameJsonString := ""
	gameJsonString += "{"
	gameJsonString += "\"gameName\": \"" + game.gameName + "\","
	gameJsonString += "\"gameId\": \"" + game.gameId + "\","
	gameJsonString += "\"roundTimer\": " + fmt.Sprint(game.roundTimer) + ","
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
		for j := range prompt {
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
		for j := range drawing {
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
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not found\"}"
		fmt.Fprintf(w, responseStr)
		return
	}
	gameStateJSON := gameStateToJSON(*game)
	fmt.Fprintf(w, gameStateJSON)
}

func getEndedGame(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	jsonObject := parseBodyObject(r)
	endedGame, ok := endedGames[jsonObject["gameId"]]
	if !ok {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not found\"}"
		fmt.Fprintf(w, responseStr)
		return
	}
	endedGameStateJSON := endedGameStateToJSON(*endedGame)
	fmt.Fprintf(w, endedGameStateJSON)
}

func startGame(w http.ResponseWriter, r *http.Request) {
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	game, ok := games[gameName]
	if !ok {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not found\"}"
		fmt.Fprintf(w, responseStr)
		return
	}

	if len(game.players) == 0 {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"No players in game\"}"
		fmt.Fprintf(w, responseStr)
	} else if game.gameStarted == false {
		if game.totalRounds <= 0 {
			game.totalRounds = len(game.players)
		}
		game.gameStarted = true
		game.prompts = make([][]string, len(game.players))
		game.drawings = make([][]string, len(game.players))
		for i := range game.prompts {
			game.prompts[i] = make([]string, game.totalRounds)
			game.drawings[i] = make([]string, game.totalRounds)
		}

		for _, p := range game.players {
			p.queuedMessage = gameStartedMessage
		}

		responseStr := "{\"status\": \"OK\", \"message\": \"Game started\"}"
		fmt.Fprintf(w, responseStr)
	} else {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Game already started\"}"
		fmt.Fprintf(w, responseStr)
	}
}

func _endGame(gameName string) {
	// TODO move the game results to a new EndedGame struct and add it to the endedGames map (slice?)
	//   EndedGame struct is different in that it has the slice of gifs, drawing urls, and captions
	// 	 But no players, spectators, timers, etc.
	game := games[gameName]
	gifs := createGifsFromGame(game)

	endedGame := EndedGame{
		gameName: game.gameName,
		gameId:   game.gameId,
		prompts:  game.prompts,
		drawings: game.drawings,
		gifs:     gifs,
	}
	endedGames[game.gameId] = &endedGame

	for _, p := range game.players {
		p.queuedMessage = gameEndedMessage + ",\"endedGameId\": " + "\"" + game.gameId + "\"}"
	}
	delete(games, gameName)
}

func endGame(w http.ResponseWriter, r *http.Request) {
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	responseStr := "{\"status\": \"OK\", \"message\": \"Game ended\"}"
	fmt.Fprintf(w, responseStr)
	_endGame(gameName)
}

func _endRound(gameName string) bool {
	game, ok := games[gameName]
	if !ok {
		return false
	}

	if game.promptsSet == false {
		// if game.currentRound == game.totalRounds-1 {
		// 	_endGame(gameName)
		// } else {
		// set queued messages for players to draw
		for i, p := range game.players {
			p.queuedMessage = drawPromptMessage + "\"prompt\": \"" + game.prompts[i][game.currentRound] + "\"}"
		}

		game.promptsSet = true
		return true
		// }
	} else {
		// set queued messages for players to caption the drawings
		for i, p := range game.players {
			p.queuedMessage = captionPromptMessage + "\"image\": \"" + game.drawings[i][game.currentRound] + "\"}"
		}

		game.currentRound++
		if game.currentRound == game.totalRounds {
			_endGame(gameName)
			return true
		}
		game.promptsSet = false
		return true
	}
}

func endRound(w http.ResponseWriter, r *http.Request) {
	// End a round
	jsonObject := parseBodyObject(r)
	gameName := jsonObject["gameName"]
	_endRound(gameName)
	// Implement logic for either serving prompts or drawings to players
	// This will be done using the round number as an offset to the player index
	// First round player 1 creates prompt 1, player 2 gets prompt 1, player 3 gets prompt 2, etc.
	responseStr := "{\"status\": \"OK\", \"message\": \"Round ended\"}"
	fmt.Fprintf(w, responseStr)
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
		players[givenPlayerName] = &Player{playerName: givenPlayerName, playerSecret: givenPlayerSecret, queuedMessage: newPlayerMessage}
		return true
	}
}

func checkAuthentication(w http.ResponseWriter, r *http.Request) {
	// Check if a player is authenticated
	jsonObject := parseBodyObject(r)
	playerName := jsonObject["playerName"]
	playerSecret := jsonObject["playerSecret"]
	// returnString := "{"
	returnString := "{\"status\": \"OK\", "
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
			player.queuedMessage = joinedGameMessage
			game.players = append(game.players, player)
			responseStr := "{\"status\": \"OK\", \"message\": \"Player joined game\"}"
			fmt.Fprintf(w, responseStr)
		} else {
			game.spectators = append(game.spectators, player)
			responseStr := "{\"status\": \"OK\", \"message\": \"Player joined game as spectator\"}"
			fmt.Fprintf(w, responseStr)
		}
	} else {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Player not authenticated\"}"
		fmt.Fprintf(w, responseStr)
		return
	}
}

func progressGameIfReady(game *Game) {
	// Progress the game by calling _endRound() if all prompts are set and promptsSet is false
	// or if promptsSet is true and all drawings for this round are submitted
	roundNumber := game.currentRound
	if game.promptsSet == false {
		for _, promptSlice := range game.prompts {
			if promptSlice[roundNumber] == "" {
				return
			}
		}
	} else {
		for _, drawingSlice := range game.drawings {
			if drawingSlice[roundNumber] == "" {
				return
			}
		}
	}
	_endRound(game.gameName)
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
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not found\"}"
			fmt.Fprintf(w, responseStr)
			return
		}

		if game.gameStarted == false {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not started\"}"
			fmt.Fprintf(w, responseStr)
			return
		}
		if game.promptsSet == true {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Prompts already set for this round\"}"
			fmt.Fprintf(w, responseStr)
			return
		}
		playerIndex := getPlayerIndex(playerName, game)
		if playerIndex == -1 {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Player not in game\"}"
			fmt.Fprintf(w, responseStr)
			return
		}

		gameRotationIndex := (playerIndex + game.currentRound) % len(game.players)
		if len(game.prompts) == 0 || len(game.prompts[gameRotationIndex]) == 0 {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Game prompts slice not initialized\"}"
			fmt.Fprintf(w, responseStr)
			return
		}
		if game.prompts[gameRotationIndex][game.currentRound] == "" {
			game.prompts[gameRotationIndex][game.currentRound] = prompt
			responseStr := "{\"status\": \"OK\", \"message\": \"Prompt submitted\"}"
			fmt.Fprintf(w, responseStr)
			progressGameIfReady(game)
		} else {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Prompt already submitted\"}"
			fmt.Fprintf(w, responseStr)
		}
	} else {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Player not authenticated\"}"
		fmt.Fprintf(w, responseStr)

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
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not found\"}"
			fmt.Fprintf(w, responseStr)
			return
		}

		if game.gameStarted == false {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Game not started\"}"
			fmt.Fprintf(w, responseStr)
			return
		}
		if game.promptsSet == false {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Prompts not yet set for this round\"}"
			fmt.Fprintf(w, responseStr)
			return
		}
		playerIndex := getPlayerIndex(playerName, game)
		if playerIndex == -1 {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Player not in game\"}"
			fmt.Fprintf(w, responseStr)
			return
		}

		gameRotationIndex := (playerIndex + game.currentRound) % len(game.players)
		if len(game.drawings) == 0 || len(game.drawings[gameRotationIndex]) == 0 {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Game drawings slice not initialized\"}"
			fmt.Fprintf(w, responseStr)
			return
		}
		if game.drawings[gameRotationIndex][game.currentRound] == "" {
			game.drawings[gameRotationIndex][game.currentRound] = drawing
			responseStr := "{\"status\": \"OK\", \"message\": \"Drawing submitted\"}"
			fmt.Fprintf(w, responseStr)
			progressGameIfReady(game)

		} else {
			responseStr := "{\"status\": \"ERROR\", \"message\": \"Drawing already submitted\"}"
			fmt.Fprintf(w, responseStr)
		}
	} else {
		responseStr := "{\"status\": \"ERROR\", \"message\": \"Player not authenticated\"}"
		fmt.Fprintf(w, responseStr)
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
	responseStr := "{\"status\": \"OK\", \"message\": \"Image uploaded\", \"imageUrl\": \"" + imageUrl + "\"}"
	fmt.Fprintf(w, responseStr)
}

func parseImagePathFromUrl(inputUrl string) string {
	// given  a URL like http://localhost:8080/images/12345678.png it should return "images/12345678.png"
	returnString := ""
	finalSlashIndex := strings.LastIndex(inputUrl, "/")
	if finalSlashIndex != -1 {
		returnString = imagesDir + string(os.PathSeparator) + inputUrl[finalSlashIndex+1:]
	} else {
		returnString = inputUrl
	}

	return returnString
}

func extractHashFromImagePath(imagePath string) string {
	filename := filepath.Base(imagePath)      // e.g., "12345678.png"
	ext := filepath.Ext(filename)             // e.g., ".png"
	hash := strings.TrimSuffix(filename, ext) // e.g., "12345678"
	return hash
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Helper function to convert an image to *image.Paletted
func toPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	palettedImg := image.NewPaletted(bounds, palette.Plan9)
	draw.FloydSteinberg.Draw(palettedImg, bounds, img, image.Point{})
	return palettedImg
}

func wrapText(text string, face font.Face, maxWidth int) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return lines
	}

	spaceWidth, _ := face.GlyphAdvance(' ')
	spaceWidthInt := int(spaceWidth >> 6)

	var lineWords []string
	lineWidth := 0

	for _, word := range words {
		wordWidth := calcTextWidth(word, face)
		if lineWidth+wordWidth > maxWidth && len(lineWords) > 0 {
			// Start a new line
			lines = append(lines, strings.Join(lineWords, " "))
			lineWords = []string{word}
			lineWidth = wordWidth + spaceWidthInt
		} else {
			if len(lineWords) > 0 {
				lineWidth += spaceWidthInt
			}
			lineWords = append(lineWords, word)
			lineWidth += wordWidth
		}
	}
	// Add the last line
	if len(lineWords) > 0 {
		lines = append(lines, strings.Join(lineWords, " "))
	}

	return lines
}

func calcTextWidth(text string, face font.Face) int {
	width := 0
	for _, x := range text {
		awidth, ok := face.GlyphAdvance(x)
		if ok != true {
			continue
		}
		width += int(awidth >> 6)
	}
	return width
}

func createCaptionImage(caption string) string {
	// This function creates an image with the caption text
	captionImagePath := ""
	// Create a new 1024x1024 image with a white background
	imgWidth := 1024
	imgHeight := 1024
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	white := color.White
	draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	// Load the font
	fontBytes, err := os.ReadFile("fonts/" + fontName)
	if err != nil {
		fmt.Println("Error loading font:", err)
		return ""
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println("Error parsing font:", err)
		return ""
	}

	// Initialize the context
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black) // Set text color

	// Starting font size
	fontSize := 48.0
	minFontSize := 12.0
	maxWidth := imgWidth - 40 // Allow some padding

	var lines []string
	var totalTextHeight int
	var face font.Face
	var lineHeight int

	// Adjust font size to fit text vertically
	for fontSize >= minFontSize {
		c.SetFontSize(fontSize)
		face = truetype.NewFace(f, &truetype.Options{
			Size: fontSize,
		})
		lines = wrapText(caption, face, maxWidth)
		lineHeight = c.PointToFixed(fontSize * 1.5).Ceil() // Adjust line spacing
		totalTextHeight = lineHeight * len(lines)
		if totalTextHeight <= imgHeight-40 { // Allow some vertical padding
			break
		}
		fontSize -= 2 // Decrease font size and try again
	}

	if fontSize < minFontSize {
		fmt.Println("Text is too long to fit into the image")
		return ""
	}

	// Starting vertical position
	y := (imgHeight - totalTextHeight) / 2

	// Draw each line
	for _, line := range lines {
		textWidth := calcTextWidth(line, face)
		x := (imgWidth - textWidth) / 2
		pt := freetype.Pt(x, y+int(c.PointToFixed(fontSize)>>6))
		_, err = c.DrawString(line, pt)
		if err != nil {
			fmt.Println("Error drawing text:", err)
			return ""
		}
		y += lineHeight
	}

	// Generate a short hash for the filename (implement your own generateShortHash)
	hash := generateShortHash()
	if hash == "" {
		fmt.Println("Error generating file name")
		return ""
	}

	// Ensure the images directory exists
	if _, err := os.Stat("images"); os.IsNotExist(err) {
		err = os.Mkdir("images", os.ModePerm)
		if err != nil {
			fmt.Println("Error creating images directory")
			return ""
		}
	}

	// Create the output file
	captionImagePath = fmt.Sprintf("images/%s.png", hash)
	outFile, err := os.Create(captionImagePath)
	if err != nil {
		fmt.Println("Unable to create the file for writing")
		return ""
	}
	defer outFile.Close()

	// Save the image in PNG format
	err = png.Encode(outFile, img)
	if err != nil {
		fmt.Println("Error encoding image to PNG")
		return ""
	}

	return captionImagePath
}

func renameImage(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

func getNonSubmissionImagePath(captionOrDrawing string) string {
	_imageName := nonSubmissionImageName_drawing
	_string := nonSubmissionString_drawing

	if captionOrDrawing == "caption" {
		_imageName = nonSubmissionImageName_caption
		_string = nonSubmissionString_caption
	} else if captionOrDrawing == "drawing" {
		_imageName = nonSubmissionImageName_drawing
		_string = nonSubmissionString_drawing
	} else {
		fmt.Println("Error: invalid argument for getNonSubmissionImagePath()")
		return ""
	}

	imagePath := imagesDir + string(os.PathSeparator) + _imageName
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		newImagePath := createCaptionImage(_string)
		err := renameImage(newImagePath, imagePath)
		if err != nil {
			fmt.Println("Error renaming image:", err)
			return ""
		}
	}

	return imagePath
}

func createGif(drawings, captions []string) string {
	// Ensure the number of drawings and captions match
	if len(drawings) != len(captions) {
		fmt.Println("Error: number of drawings and captions do not match")
		return ""
	}

	var gifImages gif.GIF

	for i := 0; i < len(drawings); i++ {
		// Create caption image
		caption := captions[i]
		captionImagePath := ""
		if caption == "" {
			captionImagePath = getNonSubmissionImagePath("caption")
		} else {
			captionImagePath = createCaptionImage(captions[i])
		}
		if captionImagePath == "" {
			fmt.Println("Error creating caption image")
			return ""
		}

		// Load caption image
		captionImg, err := loadImage(captionImagePath)
		if err != nil {
			fmt.Println("Error loading caption image:", err)
			return ""
		}

		// Convert to paletted image
		captionPaletted := toPaletted(captionImg)

		// Add to GIF frames with 3 seconds delay (300 units)
		gifImages.Image = append(gifImages.Image, captionPaletted)
		gifImages.Delay = append(gifImages.Delay, 300)

		// Get drawing image path
		drawingUrl := drawings[i]
		if drawingUrl == "" {
			drawingUrl = "/" + getNonSubmissionImagePath("drawing")
			// The / is added to ensure the string is parsed correctly in the next step
		}
		drawingImagePath := parseImagePathFromUrl(drawingUrl)
		if drawingImagePath == "" {
			fmt.Println("Error parsing drawing image path")
			return ""
		}

		// Load drawing image
		drawingImg, err := loadImage(drawingImagePath)
		if err != nil {
			fmt.Println("Error loading drawing image:", err)
			return ""
		}

		// Convert to paletted image
		drawingPaletted := toPaletted(drawingImg)

		// Add to GIF frames with 5 seconds delay (500 units)
		gifImages.Image = append(gifImages.Image, drawingPaletted)
		gifImages.Delay = append(gifImages.Delay, 500)
	}

	// Ensure the gifs directory exists
	if _, err := os.Stat("gifs"); os.IsNotExist(err) {
		err = os.Mkdir("gifs", os.ModePerm)
		if err != nil {
			fmt.Println("Error creating gifs directory:", err)
			return ""
		}
	}

	// Get the hash from the first drawing to name the GIF
	firstDrawingImagePath := parseImagePathFromUrl(drawings[0])
	gifHash := extractHashFromImagePath(firstDrawingImagePath)
	gifFilePath := fmt.Sprintf("gifs/%s.gif", gifHash)

	// Create the output file
	outFile, err := os.Create(gifFilePath)
	if err != nil {
		fmt.Println("Unable to create the gif file for writing:", err)
		return ""
	}
	defer outFile.Close()

	// Encode the GIF and write to file
	err = gif.EncodeAll(outFile, &gifImages)
	if err != nil {
		fmt.Println("Error encoding gif:", err)
		return ""
	}
	fmt.Println("GIF created successfully")
	return gifFilePath
}

func createGifsFromGame(game *Game) []string {
	// Create a GIF for each prompt drawing chain
	var gifFilePaths []string
	for i := 0; i <= len(game.prompts)-1; i++ {
		// Get the prompt chain and the drawing chain
		promptChain := game.prompts[i]
		drawingChain := game.drawings[i]

		// Create a GIF from the prompt chain
		gifFilePath := createGif(drawingChain, promptChain)
		if gifFilePath == "" {
			fmt.Println("Error creating GIF from prompt chain")
			continue
		}

		gifFilePaths = append(gifFilePaths, gifFilePath)
	}

	return gifFilePaths
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
// -An endpoint for uploading images to the server which returns a URL for the image
//
// -Functions for queueing content to display to players when requested
// 		-A map of player names to queued messages in JSON format. If the message is an image, the frontend will handle it
// -Functions for when the game starts
// -Functions for passing prompts around
// -Functions for passing drawings around
// -Function for resizing uploaded images
//
// Functions for when the game ends
// 		-Create a GIF from each prompt drawing chain
// 		-Game state should be transfered to a new map of completed games, accessible by a different endpoint
// -Function for creating GIFs from images and captions
//
// Routine for automatically progressing the game based on a configurable timer

func main() {
	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("/listGames", listGames)
	http.HandleFunc("/getGameState", getGameState)
	http.HandleFunc("/getEndedGame", getEndedGame)
	http.HandleFunc("/startGame", startGame)
	http.HandleFunc("/endGame", endGame)
	http.HandleFunc("/endRound", endRound)
	http.HandleFunc("/checkAuthentication", checkAuthentication)
	http.HandleFunc("/joinGame", joinGame)
	http.HandleFunc("/submitPrompt", submitPrompt)
	http.HandleFunc("/submitDrawing", submitDrawing)
	http.HandleFunc("/uploadDrawing", uploadDrawing)
	http.HandleFunc("/getPlayerMessage", getPlayerQueuedMessage)

	is := http.FileServer(http.Dir("images"))
	http.Handle("/images/", http.StripPrefix("/images/", is))
	// example: http://localhost:9119/images/12345678.png
	gs := http.FileServer(http.Dir("images"))
	http.Handle("/gifs/", http.StripPrefix("/gifs/", gs))
	http.ListenAndServe(":9119", nil)
}
