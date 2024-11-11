## Backend
- Implement optional round timer
  - potetnially add a secondary timer for prompts
- Response messages should include game state codes, rather than just the inclusion or exclusion of fields

- Change all of the endpoints to begin with /pt/

## Frontend
- Add a dropdown containing options for creating a game
  - recall -1 rounds sets rounds equal to the number of players
- Add a card which shows the names of players in the game
- Add a drawing canvas
- Add a slow-reveal of the prompt-drawing chains
  - this may require some changes to the backend as well
- Add list of completed game Ids
- Add an indication of when the round ends
  - Additionally, configure trigger a useEffect after the timer ends if it is set
- Interaction display should occasionally query the server for the current game state to keep up to date with the game
