curl -X POST http://localhost:9119/uploadDrawing \
	-F "playerName=player1" \
	-F "playerSecret=secret1" \
	-F "file=@image.png"
