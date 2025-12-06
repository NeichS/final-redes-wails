import { useState, useRef, useEffect } from "react";
import type { FileInfo } from "../interfaces/FileInfo.js";
import type { ProgressInfo } from "../interfaces/ProgressInfo.js";
import type { EventMessage } from "../interfaces/EventMessage.js";
import "../styles/App.css";
import { Icon } from "@iconify/react";
import {
  EventsOn,
  EventsOff,
} from "../../wailsjs/runtime/runtime.js";
import { SendFileHandler, ToggleDowntime } from "../../wailsjs/go/server/Client.js";
import {
  ReceiveFileHandler,
  StopServerHandler,
} from "../../wailsjs/go/server/Server.js";
import { SelectFile } from "../../wailsjs/go/app/App.js";
function App() {
  const [recibir, setRecibir] = useState(false);
  const [serverOn, setServerOn] = useState(false);
  const [enviando, setEnviando] = useState(false);
  const [fileInfo, setFileInfo] = useState<FileInfo>({
    address: "",
    port: "8080",
    tcp: true,
    paths: [],
  });
  const [events, setEvents] = useState<EventMessage[]>([]);
  const [progress, setProgress] = useState<ProgressInfo>({
    visible: false,
    fileName: "",
    currentFile: 0,
    totalFiles: 0,
    sent: 0,
    total: 100,
  });

  const modalRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const modal = modalRef.current;
    if (modal) {
      if (progress.visible) {
        modal.showModal();
      } else {
        modal.close();
      }
    }
  }, [progress.visible]);

  const addEvent = (text: string, type: EventMessage["type"]) => {
    const newEvent: EventMessage = {
      id: Date.now() + Math.random(),
      text,
      type,
    };
    setEvents((prevEvents) => [...prevEvents, newEvent]);
    setTimeout(() => {
      setEvents((prevEvents) => prevEvents.filter((e) => e.id !== newEvent.id));
    }, 5000);
  };

  useEffect(() => {
    EventsOn("reception-started", (fileName) =>
      addEvent(`Recibiendo archivo: ${fileName}...`, "info")
    );
    EventsOn("reception-finished", (message) => {
      addEvent(message, "success");
      setTimeout(
        () => setProgress((prev) => ({ ...prev, visible: false })),
        2000
      );
    });
    EventsOn("client-error", (message) => {
      addEvent(message, "error");
      setTimeout(
        () => setProgress((prev) => ({ ...prev, visible: false })),
        2000
      );
      setEnviando(false);
    });
    EventsOn("server-error", (message) => addEvent(message, "error"));

    EventsOn("sending-file-start", (data) => {
      setProgress({
        visible: true,
        fileName: data.fileName,
        currentFile: data.currentFile,
        totalFiles: data.totalFiles,
        sent: 0,
        total: 1,
      });
    });
    EventsOn("sending-file-progress", (data) => {
      setProgress((prev) => ({ ...prev, sent: data.sent, total: data.total }));
    });

    EventsOn("receiving-file-progress", (data) => {
      setProgress((prev) => ({
        ...prev,
        visible: true,
        sent: data.received,
        total: data.total,
      }));
    });

    return () => {
      EventsOff(
        "reception-started",
        "reception-finished",
        "client-error",
        "server-error",
        "sending-file-start",
        "sending-file-progress",
        "receiving-file-progress"
      );
    };
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "d") {
        ToggleDowntime(true);
      }
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.key === "d") {
        ToggleDowntime(false);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, []);

  const addPaths = (incoming: string[]) => {
    setFileInfo((prev) => ({
      ...prev,
      paths: Array.from(new Set([...prev.paths, ...incoming])),
    }));
  };

  const limpiarPaths = () => setFileInfo((prev) => ({ ...prev, paths: [] }));

  const abrirFilePicker = async () => {
    try {
      const paths = await SelectFile();
      if (paths && paths.length > 0) addPaths(paths);
    } catch (err) {
      console.error(err);
    }
  };

  const enviar = async () => {
    if (!fileInfo.address.trim() || fileInfo.paths.length === 0) return;
    setEnviando(true);
    try {
      await SendFileHandler(fileInfo);
      limpiarPaths();
    } catch (err) {
      console.error(err);
    } finally {
      setEnviando(false);
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
      await ReceiveFileHandler();
    } catch (err) {
      console.error(err);
    } finally {
      setServerOn(false);
    }
  };

  return (
    <div
      data-theme="synthwave"
      className="flex flex-col min-h-screen bg-base-200"
    >
      <div className="toast toast-top toast-end z-50">
        {events.map((event) => (
          <div key={event.id} className={`alert alert-${event.type} shadow-lg`}>
            <span>{event.text}</span>
          </div>
        ))}
      </div>

      <dialog
        id="progress_modal"
        className="modal modal-bottom sm:modal-middle"
        ref={modalRef}
      >
        <div className="modal-box">
          <div className="w-full flex flex-col items-center gap-2">
            <h3 className="font-bold text-lg text-primary">
              {recibir ? "Recibiendo Archivos" : "Enviando Archivos"}
            </h3>
            <span className="text-secondary text-sm">
              {recibir
                ? `Fragmentos recibidos: ${progress.sent} de ${progress.total}`
                : `Archivo ${progress.currentFile} de ${progress.totalFiles}: ${progress.fileName}`}
            </span>
            {progress.arqs !== undefined && (
               <span className="text-warning text-xs font-bold">
                 ARQs (Retransmisiones): {progress.arqs}
               </span>
            )}
            {!recibir && (
              <span className="text-secondary text-xs">
                Fragmentos enviados: {progress.sent} de {progress.total}
              </span>
            )}
            <progress
              className="progress progress-primary w-full"
              value={progress.sent}
              max={progress.total}
            ></progress>
            <span className="font-mono">
              {Math.round((progress.sent / progress.total) * 100 || 0)}%
            </span>
          </div>
        </div>
      </dialog>

      <div className="container mx-auto flex-col justify-center items-center flex-1 flex gap-4 p-4">
        <h1 className="text-primary text-3xl font-bold">
          Transfiere tus Archivos
        </h1>

        <div className="tabs tabs-boxed flex gap-4">
          <a
            className={`tab ${
              !recibir ? "tab-active" : ""
            } text-secondary bg-secondary-content rounded-2xl`}
            onClick={() => recibir && handleMode()}
          >
            Transmitir
          </a>
          <a
            className={`tab ${
              recibir ? "tab-active" : ""
            } bg-secondary-content text-secondary rounded-2xl`}
            onClick={() => !recibir && handleMode()}
          >
            Recibir
          </a>
        </div>

        {recibir ? (
          <div className="flex flex-col items-center gap-4 p-8">
            <p className="label text-xl">Esperando archivos en puerto 8080</p>
            <span className="loading loading-spinner text-primary loading-lg"></span>
          </div>
        ) : (
          <div className="w-full max-w-xl flex flex-col items-center gap-4">
            <div className="self-center">
              <label
                className={`swap swap-rotate btn btn-ghost px-4 ${
                  enviando ? "btn-disabled" : ""
                }`}
                title="Cambiar protocolo (TCP/UDP)"
              >
                {/* El checkbox controla el estado de la animación y el valor TCP/UDP */}
                <input
                  type="checkbox"
                  checked={fileInfo.tcp}
                  onChange={(e) =>
                    setFileInfo((prev) => ({ ...prev, tcp: e.target.checked }))
                  }
                  disabled={enviando}
                />

                <div className="swap-on flex items-center gap-2">
                  <Icon className="text-primary" icon="mdi:lan" width="22" height="22" />
                  <span className="text-primary font-mono font-bold">TCP</span>
                </div>

                <div className="swap-off flex items-center gap-2">
                  <Icon className="text-accent" icon="mdi:lan-connect" width="22" height="22" />
                  <span className="text-accent font-mono font-bold">UDP</span>
                </div>
              </label>
            </div>

            <fieldset className="form-control w-full flex flex-col">
              <div className="self-center">
                <label className="label">
                  <legend className="fieldset-legend text-secondary text-center">
                    Dirección destino
                  </legend>
                </label>
              </div>

              <div className="join">
                <input
                  type="text"
                  className="input input-bordered join-item w-full"
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
                  className="input input-bordered join-item w-24"
                  placeholder="8080"
                  value={fileInfo.port}
                  onChange={(e) =>
                    setFileInfo((prev) => ({ ...prev, port: e.target.value }))
                  }
                />
              </div>
            </fieldset>

            {/* --- NUEVO PANEL DE SELECCIÓN DE ARCHIVOS --- */}
            <div className="w-full card bg-base-100 shadow-md">
              <div className="card-body p-4">
                <h2 className="card-title text-base justify-center text-secondary">
                  Archivos para Enviar
                </h2>
                {fileInfo.paths.length > 0 ? (
                  <ul className="w-full list-disc pl-5 text-left max-h-32 overflow-y-auto my-2">
                    {fileInfo.paths.map((p, i) => (
                      <li key={i} className="truncate text-sm" title={p}>
                        {p.split(/[\\/]/).pop()}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-base-content/70 italic my-4 text-center">
                    Ningún archivo seleccionado
                  </p>
                )}

                <div className="card-actions justify-center mt-2">
                  <button
                    className="btn btn-secondary btn-sm"
                    onClick={abrirFilePicker}
                  >
                    <Icon icon="mdi:file-plus-outline" width="20" height="20" />
                    Añadir Archivos
                  </button>
                  <button
                    className="btn btn-error btn-sm btn-outline"
                    onClick={limpiarPaths}
                    disabled={fileInfo.paths.length === 0}
                  >
                    <Icon
                      icon="mdi:file-remove-outline"
                      width="20"
                      height="20"
                    />
                    Limpiar
                  </button>
                </div>
              </div>
            </div>

            <button
              className="btn btn-primary btn-wide mt-4"
              onClick={enviar}
              disabled={
                enviando ||
                !fileInfo.address.trim() ||
                fileInfo.paths.length === 0
              }
            >
              {enviando ? (
                <span className="loading loading-spinner"></span>
              ) : (
                <Icon
                  icon="material-symbols:send-outline"
                  width="24"
                  height="24"
                />
              )}
              {enviando ? "Enviando..." : "Enviar"}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
