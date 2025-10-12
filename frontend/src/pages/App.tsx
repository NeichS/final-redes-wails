import { useState, useRef, useEffect } from "react"; // <-- Importa useRef
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

interface EventMessage {
  id: number;
  text: string;
  type: 'success' | 'error' | 'info';
}

interface ProgressInfo {
  visible: boolean;
  fileName: string;
  currentFile: number;
  totalFiles: number;
  sent: number;
  total: number;
}

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

  // --- PASO 1: Crear una referencia para el modal ---
  const modalRef = useRef<HTMLDialogElement>(null);

  // --- PASO 2: Usar useEffect para mostrar/ocultar el modal ---
  useEffect(() => {
    const modal = modalRef.current;
    if (modal) {
      if (progress.visible) {
        modal.showModal(); // Muestra el modal
      } else {
        modal.close(); // Oculta el modal
      }
    }
  }, [progress.visible]);


  const addEvent = (text: string, type: EventMessage['type']) => {
    const newEvent: EventMessage = {
      id: Date.now() + Math.random(),
      text,
      type,
    };
    setEvents(prevEvents => [...prevEvents, newEvent]);
    setTimeout(() => {
      setEvents(prevEvents => prevEvents.filter(e => e.id !== newEvent.id));
    }, 5000);
  };
  
  useEffect(() => {
    EventsOn("reception-started", (fileName) => addEvent(`Recibiendo archivo: ${fileName}...`, 'info'));
    EventsOn("reception-finished", (message) => {
        addEvent(message, 'success');
        // Oculta la barra de progreso al finalizar
        setTimeout(() => setProgress(prev => ({ ...prev, visible: false })), 2000);
    });
    EventsOn("client-error", (message) => {
        addEvent(message, 'error');
        // Oculta la barra de progreso si hay error
        setTimeout(() => setProgress(prev => ({ ...prev, visible: false })), 2000);
        setEnviando(false)
    });
    EventsOn("server-error", (message) => addEvent(message, 'error'));
    
    // Listeners para el progreso
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
      setProgress((prev) => ({
        ...prev,
        sent: data.sent,
        total: data.total,
      }));
    });

    return () => {
      EventsOff("reception-started", "reception-finished", "client-error", "server-error", "sending-file-start", "sending-file-progress");
    };
  }, []);

  const addPaths = (incoming: string[]) => {
    setFileInfo((prev) => ({ ...prev, paths: Array.from(new Set([...prev.paths, ...incoming])) }));
  };

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => addPaths(paths), true);
    return () => OnFileDropOff();
  }, []);

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
    setEnviando(true);
    if (!fileInfo.address.trim() || fileInfo.paths.length === 0) return;
    try {
      await SendFileHandler(fileInfo);
      limpiarPaths();
      setEnviando(false);
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
      await ReceiveFileHandler();
      setServerOn(false);
    } catch (err) {
      console.error(err);
      setServerOn(false);
    }
  };

  return (
    <div data-theme="synthwave" className="flex flex-col min-h-screen bg-base-200">
      <div className="toast toast-top toast-end z-50">
        {events.map((event) => (
          <div key={event.id} className={`alert alert-${event.type}`}>
            <span>{event.text}</span>
          </div>
        ))}
      </div>

      <dialog id="progress_modal" className="modal modal-bottom sm:modal-middle" ref={modalRef}>
        <div className="modal-box">
          <div className="w-full flex flex-col items-center gap-2">
            <h3 className="font-bold text-lg text-primary">Enviando Archivos</h3>
            <span className="text-secondary text-sm">
              Archivo {progress.currentFile} de {progress.totalFiles}: {progress.fileName}
            </span>
            <progress 
              className="progress progress-primary w-full" 
              value={progress.sent} 
              max={progress.total}>
            </progress>
            <span className="font-mono">
              {Math.round((progress.sent / progress.total) * 100 || 0)}%
            </span>
          </div>
        </div>
      </dialog>

      <div className="container mx-auto flex-col justify-center items-center flex-1 flex gap-4">
        <h1 className="text-primary">Transfiera o reciba su archivo</h1>
        
        {/* El resto de tu JSX queda igual, ya no necesitamos la barra de progreso aquí */}

        <div className="flex flex-col items-center gap-4">
          <div className="flex flex-row items-center justify-center gap-4">
            <span className="badge badge-lg bg-secondary-content text-secondary">
              {recibir ? "Receptor" : "Transmisor"}
            </span>
            <label className="swap swap-rotate bg-primary-content btn">
              <input type="checkbox" checked={recibir} onChange={handleMode} />
              <Icon className="swap-on text-primary" icon="ic:sharp-call-received" width="32" height="32" />
              <Icon className="swap-off text-primary" icon="tabler:arrows-exchange" width="32" height="32" />
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
                <input type="checkbox" checked={fileInfo.tcp} onChange={(e) => setFileInfo((prev) => ({ ...prev, tcp: e.target.checked }))} />
                <Icon className="swap-off text-primary" icon="tabler:cube-send" width="32" height="32" />
                <Icon className="swap-on text-primary" icon="material-symbols:handshake-outline" width="32" height="32" />
              </label>
            </div>
            <fieldset className="fieldset flex flex-col items-center justify-center text-center">
              <legend className="fieldset-legend text-secondary text-center">Dirección destino</legend>
              <div className="flex flex-row items-center justify-center">
                <input type="text" className="input" placeholder="127.0.0.1" value={fileInfo.address} onChange={(e) => setFileInfo((prev) => ({ ...prev, address: e.target.value }))} />
                <input type="number" className="input flex-1" placeholder="8080" maxLength={5} value={fileInfo.port} onChange={(e) => setFileInfo((prev) => ({ ...prev, port: e.target.value }))} />
              </div>
              <p className="label">Ingrese la dirección del host a conectarse</p>
            </fieldset>
            <button className="w-full max-w-xl border-2 border-dashed rounded-xl p-8 flex flex-col items-center text-center gap-2 transition bg-base-100" onClick={abrirFilePicker} style={{ "--wails-drop-target": "drop" } as React.CSSProperties}>
              <p className="text-base-content/70">Arrastre aqui o clickee para examinar</p>
              <Icon icon="mdi:file-upload-outline" width="48" height="48" className="" />
            </button>
            <button type="button" className="btn btn-error" onClick={limpiarPaths}>
              <Icon icon="mdi:file-remove-outline" width="20" height="20" /> Limpiar
            </button>
            <ul className="w-full max-w-xl list-disc pl-6 text-center">
              {fileInfo.paths.map((p, i) => ( <li key={i} className="truncate">{p}</li> ))}
            </ul>
            <div className="flex gap-2">
              {!enviando && (
                <button className="btn btn-primary text-primary-content" onClick={enviar} disabled={!fileInfo.address.trim() || fileInfo.paths.length === 0}>
                Enviar <Icon icon="material-symbols:send-outline" width="24" height="24" />
              </button>
              )}
              
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;