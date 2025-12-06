# Documentación del Proyecto: Sistema de Transferencia de Archivos (TCP/UDP)

## 1. Descripción General

Esta aplicación de escritorio permite la transferencia de múltiples archivos entre computadoras conectadas a una misma red de área local (LAN). Desarrollada utilizando la arquitectura **Wails** (Go para el backend y React para el frontend), la herramienta permite al usuario seleccionar dinámicamente entre un transporte fiable (**TCP**) y uno no fiable (**UDP**), proporcionando métricas en tiempo real y mecanismos de validación de integridad.

## 2. Características Principales

### A. Soporte Multi-Protocolo (TCP / UDP)

La aplicación implementa dos estrategias de transmisión a nivel de capa de transporte, seleccionables por el usuario mediante un "switch" en la interfaz:

* **Modo TCP (Fiabilidad):** Utiliza el protocolo TCP para garantizar la entrega ordenada y sin errores. Implementa un mecanismo de **ARQ (Automatic Repeat Request)** tipo *Stop-and-Wait* en la capa de aplicación, donde el cliente espera un `ACK` del servidor por cada fragmento enviado. Si no se recibe respuesta, el paquete se retransmite.
* **Modo UDP (Velocidad/Best-Effort):** Utiliza el protocolo UDP para el envío de ráfagas de datagramas. Prioriza la velocidad sobre la fiabilidad, operando bajo la lógica *fire-and-forget*, ideal para demostrar las diferencias de pérdida de paquetes en redes congestionadas.

### B. Transferencia de Múltiples Archivos

El sistema permite la selección y encolado de múltiples archivos en una sola sesión. La aplicación procesa la lista de archivos (`paths`) secuencialmente, iniciando una nueva negociación de conexión (handshake) para cada archivo automáticamente.

### C. Monitoreo y Progreso en Tiempo Real

Para brindar feedback visual del estado de la red, la interfaz incluye:

* **Barra de Progreso:** Visualización porcentual del avance del archivo actual.
* **Contador de Fragmentos:** Muestra en tiempo real la cantidad de *chunks* (fragmentos de 1024 bytes) enviados exitosamente frente al total calculado.

### D. Simulación de Interrupción de Red (Tecla 'D')

La aplicación incluye una herramienta de depuración y docencia integrada. Al presionar la tecla **`d`** durante una transmisión, se activa el modo **"Downtime"**.

* **Funcionamiento:** El cliente entra en un bucle de espera (`sleep`), pausando el envío de nuevos paquetes sin cerrar la conexión.
* **Utilidad:** Permite observar el comportamiento de los timeouts del socket y la ventana de congestión en herramientas de análisis como Wireshark.

### E. Validación de Integridad (Checksum MD5)

Para asegurar que el archivo recibido es idéntico al enviado (especialmente crítico en UDP o redes ruidosas), se implementa verificación por hash:

1. **Emisor:** Calcula el hash **MD5** del archivo antes de la transmisión.
2. **Protocolo:** Envía el hash como parte de la cabecera (Header) inicial del archivo.
3. **Receptor:** Al finalizar la recepción, recalcula el hash del archivo reconstruido y lo compara con el hash recibido en la cabecera.
4. **Resultado:** Notifica visualmente al usuario con "Éxito" o "Error de integridad".

---

## 3. Especificaciones Técnicas del Protocolo de Aplicación

La aplicación no envía el archivo "crudo", sino que encapsula los datos en una estructura de paquete personalizada para manejar la metadata.

### Estructura del Paquete (Header Inicial)

Antes de enviar el contenido, se envía un paquete de control (Tipo 1) que contiene:

1. **Tipo de Mensaje:** Byte identificador (1 = Inicio).
2. **Total Reps:** Cantidad total de fragmentos en los que se dividirá el archivo.
3. **Longitud Nombre:** Largo del nombre del archivo.
4. **Longitud Checksum:** Largo del string MD5.
5. **Payload:** Nombre del archivo + Checksum MD5.

### Estructura del Fragmento de Datos

El archivo se divide en chunks de **1024 bytes** (payload efectivo) + cabeceras:

1. **Tipo de Mensaje:** Byte identificador (2 = Datos).
2. **Número de Secuencia (Seq):** `uint32` para ordenar los paquetes en el receptor.
3. **Longitud de Datos:** Cantidad de bytes útiles en este paquete.
4. **Payload:** Los bytes del archivo.

---

## 4. Guía de Uso Rápido

1. **Selección de Rol:**
    * En una PC, seleccionar la pestaña **"Recibir"**. Esta actuará como Servidor y quedará a la escucha en el puerto 8080.
    * En la otra PC, seleccionar **"Transmitir"**.
2. **Configuración del Transmisor:**
    * Ingresar la **Dirección IP** de la PC receptora.
    * Seleccionar el protocolo deseado (**TCP** o **UDP**) usando el interruptor.
3. **Selección de Archivos:**
    * Hacer clic en "Añadir Archivos" y seleccionar uno o varios documentos.
4. **Envío:**
    * Presionar "Enviar". Se desplegará el modal de progreso.
5. **Interacción:**
    * Mantener presionada la tecla **`d`** para simular una caída de enlace y ver cómo reacciona la barra de progreso (se detiene) y cómo se recupera al soltarla.
