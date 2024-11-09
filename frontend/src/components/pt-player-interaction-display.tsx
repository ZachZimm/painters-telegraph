import React, { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
// import refresh icon
import { RefreshCw } from "lucide-react";

interface PlayerInteractionDisplayProps {
  userId: string;
  userName: string;
  gameName: string;
  extShouldUpdate: boolean;
}

export function PlayerInteractionDisplay({
  userId,
  userName,
  gameName,
  extShouldUpdate,
}: PlayerInteractionDisplayProps) {
  const [playerMessageObject, setPlayerMessageObject] = useState({});
  const [gamesList, setGamesList] = useState([]);
  const [endedGamesList, setEndedGamesList] = useState([]);
  const [shouldUpdate, setShouldUpdate] = useState(false);
  const [gameState, setGameState] = useState({});
  const [gameId, setGameId] = useState("");
  const [gameEnded, setGameEnded] = useState(false);

  const drawingUploadRef = useRef<HTMLInputElement>(null);

  const fetchPlayerMessage = () => {
    const url = "http://lab-ts:9119/getPlayerMessage";
    var requestBody = {
      playerName: userName,
      playerSecret: userId,
    };
    fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    })
      .then((response) => response.json())
      .then((data) => {
        setPlayerMessageObject(data);
      });
  };

  const fetchGameData = () => {
    const url = "http://lab-ts:9119/getGameState";
    var requestBody = {
      gameName: gameName,
    };
    fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.status !== "ERROR") {
          setGameId(data.gameId);
        }
        if (parseInt(data.totalRounds) === 0) {
          setGameEnded(true);
          fetchEndedGameData();
        } else {
          setGameEnded(false);
          setGameState(data);
        }
      });
  };

  const fetchEndedGameData = () => {
    const url = "http://lab-ts:9119/getEndedGame";
    var requestBody = {
      gameId: "c5711edd5aac63d93761ca8755ba8ea5",
      // gameId: gameId,
    };
    fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    })
      .then((response) => response.json())
      .then((data) => {
        setGameState(data);
      });
  };

  const fetchEndedGamesList = () => {
    const url = "http://lab-ts:9119/listEndedGames";
    fetch(url)
      .then((response) => response.json())
      .then((data) => {
        setEndedGamesList(data.games);
      });
  };

  const fetchGamesList = () => {
    const url = "http://lab-ts:9119/listGames";
    fetch(url)
      .then((response) => response.json())
      .then((data) => {
        setGamesList(data.games);
      });
  };

  const submitPrompt = () => {
    const url = "http://lab-ts:9119/submitPrompt";
    var requestBody = {
      gameName: gameName,
      playerName: userName,
      playerSecret: userId,
      prompt: document.getElementById("promptInput").value || "",
    };
    fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    })
      .then((response) => response.json())
      .then((data) => {
        setShouldUpdate(!shouldUpdate);
      });
  };

  const submitDrawing = (drawingUrl: string) => {
    const url = "http://lab-ts:9119/submitDrawing";

    var requestBody = {
      gameName: document.getElementById("gameInput").value || "defaultGameName",
      playerName: userName,
      playerSecret: userId,
      drawing: drawingUrl,
    };
    fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    })
      .then((response) => response.json())
      .then((data) => {
        setShouldUpdate(!shouldUpdate);
      });
  };

  const uploadDrawing = () => {
    const url = "http://lab-ts:9119/uploadDrawing";

    if (!drawingUploadRef.current) {
      console.error("File input element not found");
      return;
    }

    const file = drawingUploadRef.current.files?.[0];
    if (!file) {
      alert("Please select a file before uploading.");
      return;
    }

    const formData = new FormData();
    formData.append("playerName", userName);
    formData.append("playerSecret", userId);
    formData.append("file", file);

    fetch(url, {
      method: "POST", // Use "POST" as per your requirement
      body: formData,
    })
      .then((response) => response.json())
      .then((data) => {
        submitDrawing(data.imageUrl);
      })
      .catch((error) => {
        console.error("Error uploading file:", error);
      });
  };

  useEffect(() => {
    fetchPlayerMessage();
    fetchGamesList();
    fetchEndedGamesList();
    fetchGameData();
  }, [extShouldUpdate, shouldUpdate, userName]);

  const handleKeyPress: React.KeyboardEventHandler<HTMLInputElement> = (e) => {
    if (e.key === "Enter") {
      e.preventDefault(); // Prevents the default newline insertion behavior
    }
  };
  return (
    <div className="flex flex-col h-full items-center p-3 bg-accent">
      <RefreshCw
        className="w-6 h-6 text-accent-foreground cursor-pointer ml-auto mr-2"
        onClick={() => setShouldUpdate(!shouldUpdate)}
      />
      <ScrollArea className="flex-1">
        <div className="flex flex-col items-center">
          <div>
            <h2 className="font-semibold pb-2 text-accent-foreground">
              {playerMessageObject.message}
              {parseInt(gameState.totalRounds || 0) !== 0 &&
                gameState.gameStarted == true &&
                (
                  " - Round " +
                  (parseInt(gameState.currentRound || 0) + 1)
                ).toString() +
                  " / " +
                  parseInt(gameState.totalRounds || 0).toString()}
            </h2>
          </div>
          {playerMessageObject.prompt && (
            <div>
              <h3 className="font-semibold pb-2 text-accent-foreground">
                Prompt: {playerMessageObject.prompt}
              </h3>
              <div className="grid w-full max-w-sm items-center gap-1.5">
                <Input id="drawingUpload" type="file" ref={drawingUploadRef} />
                <Button variant="default" onMouseUp={uploadDrawing}>
                  Upload
                </Button>
              </div>
              <br />
            </div>
          )}
          {playerMessageObject.image && (
            <Card className="max-w-[65%] p-6 rounded-lg mx-auto">
              <img
                src={playerMessageObject.image}
                alt="prompt image"
                className="max-w-full h-auto"
              />
            </Card>
          )}
          {(playerMessageObject.image || playerMessageObject.startPrompt) && (
            <div className="flex gap-1 mt-4">
              <Input
                placeholder="write your prompt here..."
                id="promptInput"
                className="w-[16.25rem]"
                type="text"
                onKeyDown={handleKeyPress}
              />
              <Button variant="default" onMouseUp={submitPrompt}>
                Submit
              </Button>
            </div>
          )}
          {gameId !== "" && (
            <div className="flex flex-col mt-4 max-w-[50%]">
              <Button variant="default" onMouseUp={fetchEndedGameData}>
                Fetch Ended Game Data
              </Button>
              {/* <h2 className="font-semibold flex-wrap pb-2 text-accent-foreground">
                {JSON.stringify(gameState)}
                <br />
                {gameState.gameStarted}
                <br />
                Game ID: {gameId}
              </h2> */}
            </div>
          )}
          {gameState.gifs && (
            <div className="flex flex-col gap-1 mt-4 p-6 mx-auto">
              {/* <h2 className="font-semibold pb-2 text-accent-foreground">
                Gifs:
              </h2> */}
              <Label
                className="text-accent-foreground font-semibold mr-[33%]"
                htmlFor="gifsDiv"
              >
                Gifs:
              </Label>
              <div id="gifsDiv" className="flex flex-col gap-2">
                {gameState.gifs.map((gif: string, index: number) => (
                  <Card
                    key={index}
                    className="max-w-[65%] p-6 rounded-lg mx-auto"
                  >
                    <img src={gif} alt="gif" className="max-w-full h-auto" />
                  </Card>
                ))}
              </div>
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
export default PlayerInteractionDisplay;
