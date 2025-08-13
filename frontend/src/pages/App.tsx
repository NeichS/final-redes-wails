import { useState, useRef } from "react";
import "../styles/App.css";
import { Greet, Testing } from "../../wailsjs/go/main/App.js";
import { Icon } from "@iconify/react";

function App() {
  const [recibir, setRecibir] = useState(false); //si recibir = false, es transmision
  const [tcp, setTcp] = useState(false); //si tcp = false, es udp
  const fileInputRef = useRef<HTMLInputElement>(null);

  const limpiarArchivos = () => {
    if (fileInputRef.current) {
      fileInputRef.current.value = ""; // limpia la selección
    }
  };

  const switchTcpUdp = () => {
    setTcp((prev) => !prev);
  };

  return (
    <div
      data-theme="synthwave"
      className="flex flex-col min-h-screen bg-base-200"
    >
      <div className="container mx-auto flex-col justify-center items-center flex-1 flex gap-4">
        <h1 className="text-primary">Transfiera o reciba su archivo</h1>
        <div className="flex flex-col gap-4">
          <div className="flex-1 flex flex-row justify-center items-center gap-4">
            <span className="badge badge-lg bg-secondary-content text-secondary">
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
                icon="tabler:arrows-exchange"
                width="32"
                height="32"
              />
            </label>
          </div>

          {recibir ? (
            <div className="flex flex-row justify-center items-center gap-4">
              <p className="label">Esperando archivo en puerto 8080</p>
              <span className="loading loading-spinner text-primary"></span>
            </div>
          ) : (
            <div className="flex gap-4 flex-col justify-center items-center">
              <fieldset className="fieldset">
                <legend className="fieldset-legend text-secondary">
                  Direccion host
                </legend>
                <input
                  type="text"
                  className="input"
                  placeholder="0.0.0.0:0000"
                />
                <p className="label">
                  Ingrese la direccion del host a conectarse
                </p>
              </fieldset>
              <div className="flex flex-row justify-center items-center gap-4">
                <input
                  type="file"
                  ref={fileInputRef}
                  multiple
                  className="file-input file-input-sm file-input-primary"
                />
                <button
                  type="button"
                  onClick={limpiarArchivos}
                  className="btn btn-error btn-rounded btn-sm"
                >
                  <Icon icon="mdi:file-remove-outline" width="24" height="24" />
                </button>
              </div>

              <div className="flex-1 flex flex-row justify-center items-center gap-4">
                <span className="badge badge-lg bg-secondary-content text-secondary">
                  {tcp ? "TCP" : "UDP"}
                </span>
                <label className="swap swap-rotate bg-primary-content btn">
                  {/* this hidden checkbox controls the state */}
                  <input type="checkbox" onClick={switchTcpUdp} />
                  <Icon
                    className="swap-on text-primary"
                    icon="tabler:cube-send"
                    width="32"
                    height="32"
                  />
                  <Icon
                    className="swap-off text-primary"
                    icon="material-symbols:handshake-outline"
                    width="32"
                    height="32"
                  />
                </label>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
