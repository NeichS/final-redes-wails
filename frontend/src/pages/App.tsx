import { useState, useRef, useEffect } from "react";
import "../styles/App.css";
import { Icon } from "@iconify/react";
import { OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime.js";
import { SendFile } from "../../wailsjs/go/server/FileServer.js";

interface FileInfo {
  address: string;
  port: string;
  tcp: boolean;
  paths: string[];
}

function App() {
  const [recibir, setRecibir] = useState(false); // false=Transmitir, true=Recibir
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [fileInfo, setFileInfo] = useState<FileInfo>({
    address: "",
    port: "8080",
    tcp: true,
    paths: [],
  });

  // util: evitar duplicados si el usuario arrastra varias veces lo mismo
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

  const abrirFilePicker = () => fileInputRef.current?.click();

  const onFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const list = e.target.files
      ? Array.from(e.target.files).map((f) => (f as any).path || f.name)
      : [];
    addPaths(list);
    // opcional: limpiar el input para poder volver a elegir lo mismo después
    e.target.value = "";
  };

  const enviar = async () => {
    if (!fileInfo.address.trim() || fileInfo.paths.length === 0) return;
    try {
      // Llama a tu backend Go que hace streaming por rutas
      await SendFile(fileInfo);
      // feedback al usuario…
      // toast ok, limpiar paths, etc.
      limpiarPaths();
    } catch (err) {
      console.error(err);
      // toast de error
    }
  };

  return (
    <div
      data-theme="synthwave"
      className="flex flex-col min-h-screen bg-base-200"
    >
      <div className="container mx-auto flex-col justify-center items-center flex-1 flex gap-4">
        <h1 className="text-primary">Transfiera o reciba su archivo</h1>

        {/* Selector de modo */}
        <div className="flex flex-col items-center gap-4">
          <div className="flex flex-row items-center justify-center gap-4">
            <span className="badge badge-lg bg-secondary-content text-secondary">
              {recibir ? "Receptor" : "Transmisor"}
            </span>
            <label className="swap swap-rotate bg-primary-content btn">
              <input
                type="checkbox"
                checked={recibir}
                onChange={(e) => setRecibir(e.target.checked)}
              />
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
            <p className="label">Esperando archivo en puerto 8080</p>
            <span className="loading loading-spinner text-primary"></span>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-4">
            {/* TCP/UDP */}
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
            {/* Dirección */}

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

            {/* Zona de DROP (visible) */}
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
              {/* input oculto para file picker */}
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={onFileInputChange}
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

            {/* Lista seleccionada */}
            <ul className="w-full max-w-xl list-disc pl-6 text-center">
              {fileInfo.paths.map((p, i) => (
                <li key={i} className="truncate">
                  {p}
                </li>
              ))}
            </ul>

            {/* Acciones */}
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
