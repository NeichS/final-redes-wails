package server

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// startUDPServer es el bucle principal del servidor UDP
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
	defer conn.Close()
	log.Println("Servidor UDP escuchando en :8080")

	var currentFile *os.File
	var expectedSeqNum uint32
	var totalSegments uint32
	var receivedChecksum string
	var fileName string

	buffer := make([]byte, 2048)

	for {
		s.mu.Lock()
		if !s.isListening {
			s.mu.Unlock()
			log.Println("Servidor UDP detenido.")
			return
		}
		s.mu.Unlock()
		
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		packetType := buffer[0]

		switch packetType {
		case 1:
			totalSegments = binary.BigEndian.Uint32(buffer[1:5])
			nameLen := binary.BigEndian.Uint32(buffer[5:9])
			checksumLen := binary.BigEndian.Uint32(buffer[9:13])
			
			endOfNames := 13 + nameLen
			fileName = string(buffer[13:endOfNames])
			receivedChecksum = string(buffer[endOfNames : endOfNames+checksumLen])
			
			log.Printf("UDP: Recibiendo '%s', %d segmentos, checksum %s", fileName, totalSegments, receivedChecksum)
			runtime.EventsEmit(s.ctx, "reception-started", fileName)
			
			os.MkdirAll("./receive", 0755)
			currentFile, err = os.Create("./receive/" + fileName)
			if err != nil {
				log.Printf("UDP Error al crear archivo: %v", err)
				continue
			}
			
			expectedSeqNum = 1
			sendAck(conn, clientAddr, 0) // ACK para el paquete de inicio

		case 2: // Paquete de DATOS
			if currentFile == nil { continue }
			
			seqNum := binary.BigEndian.Uint32(buffer[1:5])
			if seqNum == expectedSeqNum {
				_, err := currentFile.Write(buffer[5:n])
				if err != nil {
					log.Printf("UDP Error al escribir en archivo: %v", err)
					continue
				}
				expectedSeqNum++
			}
			sendAck(conn, clientAddr, seqNum) // Enviar ACK aunque sea un duplicado
		
		case 3: // Paquete de FIN
			if currentFile == nil { continue }
			
			sendAck(conn, clientAddr, totalSegments+1)
			currentFile.Close()
			log.Printf("UDP: Transferencia de '%s' finalizada.", fileName)
			
			// Verificar Checksum
			verifyUDPChecksum(s.ctx, fileName, receivedChecksum)
			
			// Reiniciar estado para el próximo archivo
			currentFile = nil
		}
	}
}

func sendAck(conn *net.UDPConn, addr *net.UDPAddr, ackNum uint32) {
	ackPacket := []byte("ACK")
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, ackNum)
	ackPacket = append(ackPacket, temp...)
	conn.WriteToUDP(ackPacket, addr)
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