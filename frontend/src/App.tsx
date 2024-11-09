import { useState, useEffect } from "react";
// Keeping these imports around as examples and reminders for now
import reactLogo from "./assets/react.svg";
import viteLogo from "/vite.svg";
import "./App.css";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { PlayerMessageDisplay } from "./components/pt-player-message-display";
import { PlayerInteractionDisplay } from "./components/pt-player-interaction-display";
import GithubLoginButton from "./components/github-login-button";
import ModeToggle from "./components/ui/mode-toggle";
import { Button } from "./components/ui/button";
import { Input } from "./components/ui/input";

function App() {
  const [userName, setUserName] = useState("not logged in");
  const [userId, setUserId] = useState("");
  const [nameInputValue, setNameInputValue] = useState("");
  const [gameInputValue, setGameInputValue] = useState("");
  const [shouldUpdate, setShouldUpdate] = useState(false);
  const [displayName1, setDisplayName1] = useState("pt-playerMessage");
  const [displayName2, setDisplayName2] = useState("pt-playerInteraction");
  const [displayName3, setDisplayName3] = useState("portfolio"); // Need to create a third display

  const setPlayerNameFunc = () => {
    setUserName(nameInputValue || "not logged in");
  };

  const handleKeyPress: React.KeyboardEventHandler<HTMLInputElement> = (e) => {
    if (e.key === "Enter") {
      e.preventDefault(); // Prevents the default newline insertion behavior
      setPlayerNameFunc();
      setShouldUpdate(!shouldUpdate);
    }
  };

  useEffect(() => {
    // Check if user is logged in
    const userData = document.cookie
      .split("; ")
      .find((row) => row.startsWith("user_id="))
      ?.split("=")[1];
    console.log(document.cookie);

    if (userData) {
      // Fetch user data from backend
      fetch(`https://host.zzimm.com/api/user/${userData}`)
        .then((res) => res.json())
        .then((data) => {
          // Data is returned as [user_id, username]
          setUserName(data[1]);
          setUserId(data[0]);
        });
    }
  }, [userName, shouldUpdate]);

  return (
    <div className="h-[92vh] w-[93vw] items-center justify-center">
      <ResizablePanelGroup
        direction="vertical"
        className="justify-self-center h-full rounded-lg border"
      >
        <ResizablePanel defaultSize={6} className="w-full">
          {/* Header
                This will eventually go into a seperate component */}
          <div className="flex flex-row flex-1 py-1 px-2">
            <div className="gap-1 p-0">
              <h2 className="font-semibold text-lg justify-self-center">
                Painter's Telegraph
                {userName !== "not logged in" && " - " + userName}
              </h2>
              <div className="flex gap-1">
                <Input
                  placeholder="game name..."
                  id="gameInput"
                  className="w-[8rem]"
                  type="text"
                  onChange={(e) => setGameInputValue(e.target.value)}
                  onKeyDown={handleKeyPress}
                />
                <Input
                  placeholder="your name..."
                  id="nameInput"
                  className="w-[8rem]"
                  type="text"
                  onChange={(e) => setNameInputValue(e.target.value)}
                  onKeyDown={handleKeyPress}
                />

                <Button variant="default" onMouseUp={setPlayerNameFunc}>
                  Submit
                </Button>
              </div>
            </div>
            {/* TODO This div should be made into a login component once we add more login methods such as other auth providers, NOSTR, username / pass, blockchain, etc... */}
            <div className="flex-1 justify-end flex gap-2">
              <GithubLoginButton username={userName} userId={userId} />
              <ModeToggle />
            </div>
          </div>
        </ResizablePanel>
        <ResizableHandle withHandle />
        <ResizablePanel className="w-full h-full">
          <ResizablePanelGroup
            direction="horizontal"
            className="justify-self-center h-full rounded-lg border"
          >
            <ResizablePanel className="w-full">
              {displayName1 === "chat" && (
                // <ChatDisplay fluxnoteUsername={userId + "-" + userName} />
                <PlayerMessageDisplay
                  userId={userId}
                  userName={userName}
                  gameName={gameInputValue}
                  onUserUpdate={() => setShouldUpdate(!shouldUpdate)}
                />
              )}
              {displayName1 === "pt-playerMessage" && (
                <PlayerMessageDisplay
                  userId={userId}
                  userName={userName}
                  gameName={gameInputValue}
                  onUserUpdate={() => setShouldUpdate(!shouldUpdate)}
                />
              )}
            </ResizablePanel>
            <ResizableHandle withHandle />
            <ResizablePanel defaultSize={70} className="w-full">
              {displayName2 === "pt-playerInteraction" && (
                <PlayerInteractionDisplay
                  userId={userId}
                  userName={userName}
                  gameName={gameInputValue}
                  extShouldUpdate={shouldUpdate}
                />
              )}
            </ResizablePanel>
          </ResizablePanelGroup>
        </ResizablePanel>
        <ResizableHandle withHandle />
        <ResizablePanel defaultSize={0} className="w-full">
          <div className="flex flex-col flex-1 py-1 px-2">
            {displayName3 === "portfolio" && <div>Not implemented yet</div>}
          </div>
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}

export default App;
