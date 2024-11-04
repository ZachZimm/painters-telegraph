# POST localhost:9119/joinGame with gameName=test, playerName=player1, playerSecret=secret1
curl -X POST -H "Content-Type: application/json" -d '{"gameName":"test","playerName":"player1","playerSecret":"secret1","drawing":"http://localhost:9119/images/c131a983a3cc501b.png"}' http://localhost:9119/submitDrawing
