import { useState } from "react";
import "../styles/App.css";
import { Greet, Testing } from "../../wailsjs/go/main/App";
import { Icon } from "@iconify/react";

function App() {
  const [resultText, setResultText] = useState(
    "Please enter your name below 👇"
  );
  const [name, setName] = useState("");
  const [msg, setMsg] = useState("Not tested yet");
  const [modo, setModo] = useState("Transmitir"); // Estado del switch

  const updateName = (e) => setName(e.target.value);
  const updateResultText = (result) => setResultText(result);

  function greet() {
    Greet(name).then(updateResultText);
    Testing().then(setMsg);
  }

  return (
    <div id="App">
      <h1 className="text-secondary">Transfiera o reciba su archivo</h1>

      <div className="switch-container">
        <label>
          <input
            type="radio"
            value="Recibir"
            checked={modo === "Recibir"}
            onChange={(e) => setModo(e.target.value)}
          />
          Recibir
          <Icon icon="ic:baseline-call-received" width="24" height="24" />
        </label>
        <label>
          <input
            type="radio"
            value="Transmitir"
            checked={modo === "Transmitir"}
            onChange={(e) => setModo(e.target.value)}
          />
          Transmitir
          <Icon icon="ic:baseline-send" width="24" height="24" />
        </label>
      </div>
      {modo === "Transmitir" ? (
        <div>
          <div className="mode-header">
            <h3>Modo</h3>
            <h3 className="mode-text">Transmisor</h3>
          </div>
          <p>Seleccione el archivo que desea enviar.</p>
          <input type="file" />
        </div>
      ) : (
        <div>
          <div className="mode-header">
            <h3>Modo</h3>
            <h3 className="mode-text">Receptor</h3>
          </div>
          <p>Esperando conexión para recibir archivo...</p>
          <button className="btn btn-primary">Iniciar recepción</button>
        </div>
      )}
    </div>
  );
}

export default App;
