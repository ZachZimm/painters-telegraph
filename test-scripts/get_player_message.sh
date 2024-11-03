# POST localhost:9119/joinGame with gameName=test, playerName=player1, playerSecret=secret1
curl -X POST -H "Content-Type: application/json" -d '{"playerName":"player1","playerSecret":"secret1"}' http://localhost:9119/getPlayerMessage
