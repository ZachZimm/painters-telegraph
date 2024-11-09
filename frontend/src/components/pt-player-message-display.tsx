import React, { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@radix-ui/react-dropdown-menu";
import { Card } from "@/components/ui/card";
import { on } from "events";

interface PlayerMessageDisplayProps {
  userId: string;
  userName: string;
  gameName: string;
  onUserUpdate: () => void;
}

export function PlayerMessageDisplay({
  userId,
  userName,
  gameName,
  onUserUpdate,
}: PlayerMessageDisplayProps) {
  const [playerMessageObject, setPlayerMessageObject] = useState({});
  const [gamesList, setGamesList] = useState([]);
  const [shouldUpdate, setShouldUpdate] = useState(false);
  const [playerName, setPlayerName] = useState(userName);

  // const drawingUploadRef = useRef<HTMLInputElement>(null);

  const fetchPlayerMessage = () => {
    const url = "http://lab-ts:9119/getPlayerMessage";
    var requestBody = {
      playerName: playerName,
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

  const fetchGamesList = () => {
    const url = "http://lab-ts:9119/listGames";
    // GET request
    fetch(url)
      .then((response) => response.json())
      .then((data) => {
        setGamesList(data.games);
      });
  };

  const createGame = () => {
    const url = "http://lab-ts:9119/createGame";
    var requestBody = {
      playerName: userName,
      playerSecret: userId,
      gameName: document.getElementById("gameInput").value || "defaultGameName",
      totalRounds: "2",
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
        onUserUpdate();
        setShouldUpdate(!shouldUpdate);
      });
  };

  const startGame = () => {
    const url = "http://lab-ts:9119/startGame";
    var requestBody = {
      playerName: userName,
      playerSecret: userId,
      gameName: document.getElementById("gameInput").value || "defaultGameName",
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
        onUserUpdate();
        setShouldUpdate(!shouldUpdate);
      });
  };

  const endGame = () => {
    const url = "http://lab-ts:9119/endGame";
    var requestBody = {
      playerName: userName,
      playerSecret: userId,
      gameName: document.getElementById("gameInput").value || "defaultGameName",
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
        onUserUpdate();
        setShouldUpdate(!shouldUpdate);
      });
  };

  const endRound = () => {
    const url = "http://lab-ts:9119/endRound";
    var requestBody = {
      playerName: userName,
      playerSecret: userId,
      gameName: document.getElementById("gameInput").value || "defaultGameName",
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

  const joinGame = () => {
    const url = "http://lab-ts:9119/joinGame";
    var requestBody = {
      gameName: gameName || "defaultGameName",
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
        onUserUpdate();
        setShouldUpdate(!shouldUpdate);
      });
  };

  useEffect(() => {
    setPlayerName(userName);
    fetchPlayerMessage();
    fetchGamesList();
    onUserUpdate();
  }, [shouldUpdate, userName, playerName]);

  return (
    <div className="flex flex-col h-full justify-items-center p-3 bg-accent items-center">
      <ScrollArea className="flex-1">
        <Card className="max-w-[auto] p-2 rounded-lg">
          <h2 className="font-semibold justify-self-start pb-2 text-secondary-foreground">
            List of games:
          </h2>
          {gamesList.map((game, index) => (
            <div key={index} className="p-2 m-2 border rounded bg-secondary">
              <h3 className="font-bold text-secondary-foreground">{game}</h3>{" "}
            </div>
          ))}
        </Card>
        <br />

        <div className="flex flex-col">
          <div>
            <Button variant="default" onClick={fetchGamesList}>
              List Games
            </Button>
            <Button variant="default" onClick={joinGame}>
              Join Game
            </Button>
          </div>
          <br />
          <div>
            <Button variant="outline" onClick={createGame}>
              Create Game
            </Button>
            <Button variant="outline" onClick={startGame}>
              Start Game
            </Button>
          </div>
          <br />
          <div>
            <Button variant="destructive" onClick={endRound}>
              End Round
            </Button>
            <Button variant="destructive" onClick={endGame}>
              End Game
            </Button>
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}
export default PlayerMessageDisplay;
