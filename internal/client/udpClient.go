package server

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	udpPacketSize = 1024
)

func startUDPClient(ctx context.Context, addr, port string, filePaths []string) error {
	serverAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("Error resolviendo UDP: %v", err))
		return err
	}

	// Usamos DialUDP, que es simple para enviar paquetes a un solo destino.
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("No se pudo conectar (UDP): %v", err))
		return err
	}
	defer conn.Close()

	totalFiles := len(filePaths)
	for i, path := range filePaths {
		runtime.EventsEmit(ctx, "sending-file-start", map[string]interface{}{
			"fileName":    filepath.Base(path),
			"currentFile": i + 1,
			"totalFiles":  totalFiles,
		})

		err := sendSingleFileUDP(ctx, path, conn)
		if err != nil {
			runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("Error enviando %s: %v", filepath.Base(path), err))
			// Continuamos con el siguiente archivo en lugar de detenernos
		}
		// Una pequeña pausa entre archivos para que el servidor pueda procesarlos.
		time.Sleep(250 * time.Millisecond)
	}

	runtime.EventsEmit(ctx, "reception-finished", "¡Todos los archivos enviados!")
	return nil
}

// sendSingleFileUDP envía todo el archivo sin esperar ACKs.
func sendSingleFileUDP(ctx context.Context, filePath string, conn *net.UDPConn) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 1. Calcular Checksum y metadatos
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	checksum := hex.EncodeToString(hash.Sum(nil))
	file.Seek(0, 0) // Rebobinar para leer el archivo

	baseName := filepath.Base(filePath)
	fileInfo, _ := file.Stat()
	totalSegments := uint32(fileInfo.Size()/udpPacketSize) + 1

	// 2. Enviar paquete de INICIO (tipo 1)
	// Protocolo: [Tipo(1)=1, TotalSegs(4), NameLen(4), ChecksumLen(4), Name(n), Checksum(n)]
	startPacket := createStartPacket(totalSegments, baseName, checksum)
	_, err = conn.Write(startPacket)
	if err != nil {
		return fmt.Errorf("falló el envío del paquete de inicio: %w", err)
	}

	// 3. Enviar todos los paquetes de DATOS (tipo 2) en ráfaga
	// Protocolo: [Tipo(1)=2, SeqNum(4), Data(n)]
	buffer := make([]byte, udpPacketSize)
	for seqNum := uint32(1); seqNum <= totalSegments; seqNum++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		dataPacket := createDataPacket(seqNum, buffer[:n])
		_, err = conn.Write(dataPacket)
		if err != nil {
			log.Printf("Error enviando segmento %d: %v", seqNum, err)
			// En este modo simple, ignoramos el error y continuamos
		}

		// Actualizamos la barra de progreso
		runtime.EventsEmit(ctx, "sending-file-progress", map[string]interface{}{
			"sent":  seqNum,
			"total": totalSegments,
		})

		time.Sleep(1 * time.Millisecond)
	}

	endPacket := createEndPacket(totalSegments + 1)
	_, err = conn.Write(endPacket)
	if err != nil {
		log.Printf("Error enviando paquete final: %v", err)
	}

	log.Printf("Envío simple de '%s' completado.", baseName)
	return nil
}

// Funciones auxiliares para crear paquetes (hacen el código más limpio)
func createStartPacket(totalSegs uint32, name, checksum string) []byte {
	packet := []byte{1}
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, totalSegs)
	packet = append(packet, temp...)
	binary.BigEndian.PutUint32(temp, uint32(len(name)))
	packet = append(packet, temp...)
	binary.BigEndian.PutUint32(temp, uint32(len(checksum)))
	packet = append(packet, temp...)
	packet = append(packet, []byte(name)...)
	packet = append(packet, []byte(checksum)...)
	return packet
}

func createDataPacket(seqNum uint32, data []byte) []byte {
	packet := []byte{2}
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, seqNum)
	packet = append(packet, temp...)
	packet = append(packet, data...)
	return packet
}

func createEndPacket(seqNum uint32) []byte {
	packet := []byte{3}
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, seqNum)
	packet = append(packet, temp...)
	return packet
}
