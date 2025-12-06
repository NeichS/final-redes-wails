package server

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type udpTransfer struct {
	fileHandle   *os.File
	checksum     string
	totalSegs    uint32
	receivedData map[uint32][]byte
}

func (s *Server) startUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Printf("Error UDP (Resolve): %v", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("Error UDP (Listen): %v", err)
		return
	}

	s.udpConn = conn
	defer conn.Close()
	log.Println("Servidor UDP (simple) escuchando en :8080")

	// Mantenemos un mapa de las transferencias activas, identificadas por el nombre del archivo
	activeTransfers := make(map[string]*udpTransfer)
	buffer := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// Si el error es por socket cerrado, salimos
			if errors.Is(err, net.ErrClosed) {
				log.Println("Servidor UDP detenido.")
				return
			}
			// Para otros errores, logueamos y seguimos (o salimos si es crítico)
			log.Printf("Error leyendo UDP: %v", err)
			// Si el server ya no escucha, salimos
			s.mu.Lock()
			if !s.isListening {
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()
			continue
		}

		packetType := buffer[0]
		packetData := buffer[:n]

		switch packetType {
		case 1: // Paquete de INICIO
			totalSegments := binary.BigEndian.Uint32(packetData[1:5])
			nameLen := binary.BigEndian.Uint32(packetData[5:9])
			checksumLen := binary.BigEndian.Uint32(packetData[9:13])

			endOfNames := 13 + nameLen
			fileName := string(packetData[13:endOfNames])
			receivedChecksum := string(packetData[endOfNames : endOfNames+checksumLen])

			log.Printf("UDP: Iniciando recepción de '%s'", fileName)
			runtime.EventsEmit(s.ctx, "reception-started", fileName)

			os.MkdirAll("./receive", 0755)
			file, err := os.Create("./receive/" + fileName)
			if err != nil {
				log.Printf("UDP Error al crear archivo: %v", err)
				continue
			}

			activeTransfers[fileName] = &udpTransfer{
				fileHandle:   file,
				checksum:     receivedChecksum,
				totalSegs:    totalSegments,
				receivedData: make(map[uint32][]byte),
			}

		case 2: // data
			seqNum := binary.BigEndian.Uint32(packetData[1:5])
			for _, transfer := range activeTransfers {
				dataCopy := make([]byte, len(packetData[5:]))
				copy(dataCopy, packetData[5:])
				transfer.receivedData[seqNum] = dataCopy
				break
			}

		case 3: // fin
			for fileName, transfer := range activeTransfers {
				log.Printf("UDP: Finalizando recepción de '%s'", fileName)
				runtime.EventsEmit(s.ctx, "reception-finished", fileName)
				keys := make([]int, 0, len(transfer.receivedData))
				for k := range transfer.receivedData {
					keys = append(keys, int(k))
				}
				sort.Ints(keys)

				for _, k := range keys {
					transfer.fileHandle.Write(transfer.receivedData[uint32(k)])
				}
				transfer.fileHandle.Close()

				verifyUDPChecksum(s.ctx, fileName, transfer.checksum)

				delete(activeTransfers, fileName)
				break
			}
		}
	}
}

func verifyUDPChecksum(ctx context.Context, fileName, receivedChecksum string) {
	file, err := os.Open("./receive/" + fileName)
	if err != nil {
		log.Printf("UDP Checksum: No se pudo abrir el archivo: %v", err)
		return
	}
	defer file.Close()

	hash := md5.New()
	io.Copy(hash, file)
	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))

	if receivedChecksum == calculatedChecksum {
		log.Println("UDP Checksum OK!")
		runtime.EventsEmit(ctx, "reception-finished", fmt.Sprintf("✅ ¡%s (UDP) recibido y verificado!", fileName))
	} else {
		log.Println("UDP CHECKSUM ERROR!")
		runtime.EventsEmit(ctx, "server-error", fmt.Sprintf("❌ Error de checksum en %s (UDP).", fileName))
	}
}
