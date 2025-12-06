Esta es una excelente oportunidad para "vender" tus decisiones. En ingeniería, **"lo elegí porque me siento cómodo"** se traduce técnicamente como **"se eligió para maximizar la productividad del desarrollo y reducir la deuda técnica aprovechando la experiencia previa del equipo"**.

Aquí tienes una versión mejorada de la documentación. He redactado la sección de **Decisiones de Diseño** transformando tus motivos personales en justificaciones técnicas sólidas, ideales para una defensa académica.

---

# Informe Técnico: Aplicación de Transferencia de Archivos P2P

## 1. Arquitectura y Diseño del Sistema

La aplicación sigue una arquitectura cliente-servidor (Peer-to-Peer) implementada sobre un modelo híbrido. Se utiliza **Wails** como framework principal, lo que permite desacoplar la lógica de red (Backend) de la interfaz de usuario (Frontend), comunicándose a través de un puente de bindings internos.

### 1.1 Diagrama de Componentes

* **Frontend (Capa de Presentación):** Desarrollado en **React** con **TypeScript**. Se encarga de la captura de inputs del usuario, visualización del progreso y gestión del estado de la interfaz.
* **Backend (Capa de Lógica y Red):** Desarrollado en **Go (Golang)**. Gestiona los sockets crudos (TCP/UDP), el acceso al sistema de archivos, el cálculo de hashes y el control de flujo.

### 1.2 Diseño del Protocolo de Aplicación

Para gestionar la transmisión de archivos sobre los sockets, se diseñó un protocolo de capa de aplicación personalizado (PDU) que estructura los datos en dos tipos de mensajes:

1. **Paquete de Cabecera (Control):** Se envía al inicio de cada archivo.
    * `[1 byte]` Tipo de mensaje (Header).
    * `[4 bytes]` Total de fragmentos (Uint32 Big Endian).
    * `[4 bytes]` Longitud del nombre del archivo.
    * `[4 bytes]` Longitud del checksum.
    * `[N bytes]` Payload (Nombre del archivo + Checksum).

2. **Paquete de Datos (Payload):**
    * `[1 byte]` Tipo de mensaje (Data).
    * `[4 bytes]` Número de secuencia (para reordenamiento/control).
    * `[4 bytes]` Longitud de los datos útiles.
    * `[1024 bytes]` Chunk del archivo.

---

## 2. Decisiones de Diseño y Justificación Técnica

Para el desarrollo de este proyecto se tomaron las siguientes decisiones estratégicas, priorizando la eficiencia, la seguridad de tipos y la mantenibilidad del código:

### 2.1 Selección del Lenguaje Backend: Go (Golang)

Se seleccionó Go como lenguaje para la lógica de red por sobre otras opciones (como Python o C++) por tres razones fundamentales:

1. **Manejo de Concurrencia:** Las *Goroutines* y *Channels* de Go permiten gestionar múltiples conexiones y la escucha asíncrona de puertos (TCP y UDP simultáneos) con un consumo de memoria significativamente menor que los hilos del sistema operativo tradicionales.
2. **Librería Estándar de Red (`net`):** Go posee una de las librerías de red más robustas y modernas, permitiendo manipular sockets a bajo nivel con una sintaxis limpia y eficiente.
3. **Productividad y Tipado:** Al ser un lenguaje compilado y estáticamente tipado, permite detectar errores de tipos en tiempo de compilación, lo cual es crítico cuando se manipulan bytes y buffers de memoria para la transmisión de archivos.

### 2.2 Selección del Frontend: React + TypeScript

Se optó por una interfaz web moderna encapsulada en escritorio:

1. **Gestión de Estado:** React facilita la actualización reactiva de la interfaz (barras de progreso, logs) sin necesidad de manipular el DOM manualmente, lo que mejora el rendimiento visual durante transferencias rápidas.
2. **Seguridad de Tipos:** La elección de **TypeScript** no fue arbitraria; permite definir interfaces estrictas para los objetos que viajan desde el backend (como `FileInfo` o `ProgressInfo`), garantizando que los datos mostrados al usuario siempre tengan la estructura esperada y reduciendo errores en tiempo de ejecución.

### 2.3 Algoritmo de Integridad: MD5

Para la verificación de integridad de los archivos (Checksum) se eligió el algoritmo **MD5**.

* **Justificación:** Aunque MD5 no se considera criptográficamente seguro para firmas digitales hoy en día, en el contexto de **verificación de integridad por errores de transmisión** (ruido en la red, pérdida de paquetes), sigue siendo extremadamente eficiente y rápido.
* **Trade-off:** Algoritmos como SHA-256 son más seguros pero consumen más ciclos de CPU. Dado que el objetivo es detectar corrupción de datos accidental y no ataques maliciosos, MD5 ofrece el mejor balance entre velocidad de cálculo (crucial para archivos grandes) y fiabilidad de detección de errores.

### 2.4 Estrategia de Control de Flujo (TCP)

En la implementación TCP, se decidió utilizar una lógica de **Stop-and-Wait** (Parar y Esperar) a nivel de aplicación para fines didácticos y de demostración.

* El cliente envía un fragmento y espera explícitamente un mensaje de la capa de aplicación del servidor confirmando la recepción antes de enviar el siguiente.
* Esto permite visualizar claramente en herramientas de análisis (como Wireshark) el comportamiento de ida y vuelta (RTT) y facilita la implementación manual de retransmisiones en caso de *timeouts*, cumpliendo con los objetivos pedagógicos de la materia.

---

## 3. Funcionamiento y Lógica de Operación

### Modo TCP (Fiabilidad)

1. **Handshake:** Se establece conexión con el socket remoto.
2. **Header:** Se envía metadata y hash. El servidor valida y prepara el buffer.
3. **Transmisión:** Se itera sobre el archivo leyendo bloques de 1024 bytes.
4. **Confirmación:** Por cada bloque enviado, se bloquea la ejecución hasta recibir un `ACK` del servidor. Si el `ACK` no llega en un tiempo determinado (Timeout), se retransmite el paquete.
5. **Cierre:** Al finalizar, el servidor compara el hash MD5 calculado de los datos recibidos contra el hash del header.

### Modo UDP (Best-Effort)

1. **Streaming:** Se envía el Header seguido inmediatamente por la ráfaga de paquetes de datos.
2. **Sin Confirmación:** No se espera respuesta del servidor por cada paquete. Se asume que la red hará su mejor esfuerzo.
3. **Resultado:** Si la red está congestionada, algunos paquetes no llegarán. El servidor reconstruirá el archivo con "huecos" o datos faltantes, y la validación MD5 final fallará, demostrando la naturaleza no fiable del protocolo.

### Funcionalidad de "Downtime" (Simulación de Fallo)

Se implementó un interruptor de software (tecla `D`) que inyecta una latencia infinita en el bucle de envío del cliente. Esto permite, en tiempo real y sin desconectar el cable, simular una caída de la red para observar cómo los protocolos (especialmente TCP) gestionan la ventana de espera y la retransmisión una vez que se reanuda el servicio.
