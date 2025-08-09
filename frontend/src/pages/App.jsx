import { useState } from "react";
import "../styles/App.css";
import { Greet, Testing } from "../../wailsjs/go/main/App";
import { Icon } from "@iconify/react";

function App() {
  const [name, setName] = useState("");
  const [recibir, setRecibir] = useState(false); //si recibir = false, es transmision

  const updateResultText = (result) => setResultText(result);

  function greet() {
    Greet(name).then(updateResultText);
    Testing().then(setMsg);
  }

  return (
    <div
      data-theme="synthwave"
      className="flex flex-col min-h-screen bg-base-200"
    >
      <div className="container mx-auto flex-col justify-center items-center flex-1 flex gap-4">
        <h1 className="text-primary">Transfiera o reciba su archivo</h1>
        <div>
          <div className="flex-1 flex flex-row justify-center items-center gap-4">
            <span className="badge badge-lg">
              {recibir ? "Receptor" : "Transmisor"}
            </span>
            <label className="swap swap-rotate bg-primary-content btn">
              {/* this hidden checkbox controls the state */}
              <input type="checkbox" onClick={() => setRecibir((v) => !v)} />
              <Icon
                className="swap-on text-primary"
                icon="ic:sharp-call-received"
                width="32"
                height="32"
              />
              <Icon
                className="swap-off text-primary"
                icon="ic:baseline-send"
                width="32"
                height="32"
              />
            </label>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
