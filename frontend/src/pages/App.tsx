import { useState, useEffect } from "react";
import "../styles/App.css";
import { Icon } from "@iconify/react";
import {
  OnFileDrop,
  OnFileDropOff,
  EventsOn,
  EventsOff,
} from "../../wailsjs/runtime/runtime.js";
import { SendFileHandler } from "../../wailsjs/go/server/Client.js";
import { ReceiveFileHandler, StopServerHandler } from "../../wailsjs/go/server/Server.js";
import { SelectFile } from "../../wailsjs/go/app/App.js";

interface FileInfo {
  address: string;
  port: string;
  tcp: boolean;
  paths: string[];
}

// Paso 1: Definir la estructura de un mensaje de evento
interface EventMessage {
  id: number;
  text: string;
  type: 'success' | 'error' | 'info';
}

function App() {
  const [recibir, setRecibir] = useState(false);
  const [serverOn, setServerOn] = useState(false);
  const [fileInfo, setFileInfo] = useState<FileInfo>({
    address: "",
    port: "8080",
    tcp: true,
    paths: [],
  });

  const [events, setEvents] = useState<EventMessage[]>([]);

  // Paso 2: Función auxiliar para añadir y quitar eventos de la cola
  const addEvent = (text: string, type: EventMessage['type']) => {
    const newEvent: EventMessage = {
      id: Date.now() + Math.random(),
      text,
      type,
    };

    // Añadimos el nuevo evento a la lista
    setEvents(prevEvents => [...prevEvents, newEvent]);

    // Programamos que el evento se elimine después de 5 segundos
    setTimeout(() => {
      setEvents(prevEvents => prevEvents.filter(e => e.id !== newEvent.id));
    }, 5000);
  };
  
  // Paso 3: Actualizar los listeners para que usen la nueva función `addEvent`
  useEffect(() => {
    EventsOn("reception-started", (fileName) => {
      addEvent(`Recibiendo archivo: ${fileName}...`, 'info');
    });

    EventsOn("reception-finished", (message) => {
      addEvent(message, 'success');
    });

    EventsOn("client-error", (message) => {
      addEvent(message, 'error');
    });

    EventsOn("server-error", (message) => {
      addEvent(message, 'error');
    });

    return () => {
      EventsOff("reception-started", "reception-finished", "client-error", "server-error");
    };
  }, []); // El array vacío asegura que esto se configure solo una vez

  // --- El resto de tus funciones se mantienen igual ---

  const addPaths = (incoming: string[]) => {
    setFileInfo((prev) => ({
      ...prev,
      paths: Array.from(new Set([...prev.paths, ...incoming])),
    }));
  };

  useEffect(() => {
    OnFileDrop((x, y, paths) => {
      console.log(x, y, "Dropped files: ", paths);
    }, true);
    return () => OnFileDropOff();
  }, []);

  const limpiarPaths = () =>
    setFileInfo((prev) => ({
      ...prev,
      paths: [],
    }));

  const abrirFilePicker = async () => {
    try {
      const paths = await SelectFile();
      if (paths && paths.length > 0) {
        addPaths(paths);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const enviar = async () => {
    if (!fileInfo.address.trim() || fileInfo.paths.length === 0) return;
    try {
      await SendFileHandler(fileInfo);
      limpiarPaths();
    } catch (err) {
      console.error(err);
    }
  };
  
  const handleMode = async () => {
    if (!recibir) {
      setRecibir(true);
      startServer();
    } else {
      setRecibir(false);
      await StopServerHandler();
    }
  };

  const startServer = async () => {
    if (serverOn) return;
    setServerOn(true);
    try {
      console.log("Iniciando servidor...");
      await ReceiveFileHandler();
      setServerOn(false);
      console.log("El servidor se detuvo o el archivo fue recibido.");
    } catch (err) {
      console.error(err);
      setServerOn(false);
    }
  };

  return (
    <div
      data-theme="synthwave"
      className="flex flex-col min-h-screen bg-base-200"
    >
      {/* Paso 4: Renderizar la cola de eventos usando el componente toast */}
      <div className="toast toast-top toast-end z-50">
        {events.map((event) => (
          <div key={event.id} className={`alert alert-${event.type}`}>
            <span>{event.text}</span>
          </div>
        ))}
      </div>

      <div className="container mx-auto flex-col justify-center items-center flex-1 flex gap-4">
        <h1 className="text-primary">Transfiera o reciba su archivo</h1>
        
        {/* El resto de tu JSX se mantiene igual */}
        <div className="flex flex-col items-center gap-4">
          <div className="flex flex-row items-center justify-center gap-4">
            <span className="badge badge-lg bg-secondary-content text-secondary">
              {recibir ? "Receptor" : "Transmisor"}
            </span>
            <label className="swap swap-rotate bg-primary-content btn">
              <input type="checkbox" checked={recibir} onChange={handleMode} />
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
        </div>

        {recibir ? (
          <div className="flex items-center gap-4">
            <p className="label">Esperando archivos en puerto 8080</p>
            <span className="loading loading-spinner text-primary"></span>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-4">
            <div className="flex items-center gap-4">
              <span className="badge badge-lg bg-secondary-content text-secondary">
                {fileInfo.tcp ? "TCP" : "UDP"}
              </span>
              <label className="swap swap-rotate bg-primary-content btn">
                <input
                  type="checkbox"
                  checked={fileInfo.tcp}
                  onChange={(e) =>
                    setFileInfo((prev) => ({ ...prev, tcp: e.target.checked }))
                  }
                />
                <Icon
                  className="swap-off text-primary"
                  icon="tabler:cube-send"
                  width="32"
                  height="32"
                />
                <Icon
                  className="swap-on text-primary"
                  icon="material-symbols:handshake-outline"
                  width="32"
                  height="32"
                />
              </label>
            </div>
            <fieldset className="fieldset flex flex-col items-center justify-center text-center">
              <legend className="fieldset-legend text-secondary text-center">
                Dirección destino
              </legend>
              <div className="flex flex-row items-center justify-center">
                <input
                  type="text"
                  className="input"
                  placeholder="127.0.0.1"
                  value={fileInfo.address}
                  onChange={(e) =>
                    setFileInfo((prev) => ({
                      ...prev,
                      address: e.target.value,
                    }))
                  }
                />
                <input
                  type="number"
                  className="input flex-1"
                  placeholder="8080"
                  maxLength={5}
                  value={fileInfo.port}
                  onChange={(e) =>
                    setFileInfo((prev) => ({ ...prev, port: e.target.value }))
                  }
                />
              </div>
              <p className="label">
                Ingrese la dirección del host a conectarse
              </p>
            </fieldset>

            <button
              className="w-full max-w-xl border-2 border-dashed rounded-xl p-8 flex flex-col items-center text-center gap-2 transition bg-base-100"
              onClick={abrirFilePicker}
              style={{ "--wails-drop-target": "drop" } as React.CSSProperties}
            >
              <p className="text-base-content/70">
                Arrastre aqui o clickee para examinar
              </p>
              <Icon
                icon="mdi:file-upload-outline"
                width="48"
                height="48"
                className=""
              />
            </button>
            <button
              type="button"
              className="btn btn-error"
              onClick={limpiarPaths}
            >
              <Icon icon="mdi:file-remove-outline" width="20" height="20" />
              Limpiar
            </button>

            <ul className="w-full max-w-xl list-disc pl-6 text-center">
              {fileInfo.paths.map((p, i) => (
                <li key={i} className="truncate">
                  {p}
                </li>
              ))}
            </ul>

            <div className="flex gap-2">
              <button
                className="btn btn-primary text-primary-content"
                onClick={enviar}
                disabled={
                  !fileInfo.address.trim() || fileInfo.paths.length === 0
                }
              >
                Enviar
                <Icon
                  icon="material-symbols:send-outline"
                  width="24"
                  height="24"
                />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;