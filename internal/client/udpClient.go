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
	udpPacketSize = 1024 // Tamaño del payload de datos
	timeout       = 2 * time.Second
	maxRetries    = 5
)

func startUDPClient(ctx context.Context, addr, port string, filePaths []string) error {
	serverAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("Error resolviendo UDP: %v", err))
		return err
	}

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
		time.Sleep(100 * time.Millisecond)

		err := sendSingleFileUDP(ctx, path, conn)
		if err != nil {
			runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("Error enviando %s: %v", filepath.Base(path), err))
			return err
		}
	}

	runtime.EventsEmit(ctx, "reception-finished", "¡Todos los archivos enviados con éxito!")
	return nil
}

func sendSingleFileUDP(ctx context.Context, filePath string, conn *net.UDPConn) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	checksum := hex.EncodeToString(hash.Sum(nil))
	file.Seek(0, 0)

	baseName := filepath.Base(filePath)
	fileInfo, _ := file.Stat()
	totalSegments := uint32(fileInfo.Size()/udpPacketSize) + 1

	startPacket := []byte{1}
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, totalSegments)
	startPacket = append(startPacket, temp...)
	binary.BigEndian.PutUint32(temp, uint32(len(baseName)))
	startPacket = append(startPacket, temp...)
	binary.BigEndian.PutUint32(temp, uint32(len(checksum)))
	startPacket = append(startPacket, temp...)
	startPacket = append(startPacket, []byte(baseName)...)
	startPacket = append(startPacket, []byte(checksum)...)

	if err := sendAndWaitForAck(conn, startPacket, 0); err != nil {
		return fmt.Errorf("el receptor no confirmó el inicio: %w", err)
	}

	buffer := make([]byte, udpPacketSize)
	for seqNum := uint32(1); seqNum <= totalSegments; seqNum++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// Protocolo de Datos: [Tipo(1 byte)=2, SeqNum(4), Data(n)]
		dataPacket := []byte{2}
		binary.BigEndian.PutUint32(temp, seqNum)
		dataPacket = append(dataPacket, temp...)
		dataPacket = append(dataPacket, buffer[:n]...)

		if err := sendAndWaitForAck(conn, dataPacket, seqNum); err != nil {
			return fmt.Errorf("falló el envío del segmento %d: %w", seqNum, err)
		}

		runtime.EventsEmit(ctx, "sending-file-progress", map[string]interface{}{
			"sent":  seqNum,
			"total": totalSegments,
		})
	}

	endPacket := []byte{3}
	if err := sendAndWaitForAck(conn, endPacket, totalSegments+1); err != nil {
		return fmt.Errorf("el receptor no confirmó el final: %w", err)
	}

	return nil
}

func sendAndWaitForAck(conn *net.UDPConn, packet []byte, expectedAckNum uint32) error {
	ackBuffer := make([]byte, 8) // [ACK(4 bytes), SeqNum(4 bytes)]
	ackString := "ACK"

	for i := 0; i < maxRetries; i++ {
		// Enviar paquete
		if _, err := conn.Write(packet); err != nil {
			return err
		}

		// Esperar ACK con timeout
		conn.SetReadDeadline(time.Now().Add(timeout))
		n, err := conn.Read(ackBuffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Timeout esperando ACK para %d, reintentando (%d/%d)...", expectedAckNum, i+1, maxRetries)
				continue // Reintentar
			}
			return err
		}

		// Validar ACK
		if n == 8 && string(ackBuffer[:len(ackString)]) == ackString {
			receivedAckNum := binary.BigEndian.Uint32(ackBuffer[4:])
			if receivedAckNum == expectedAckNum {
				return nil // Éxito
			}
		}
	}
	return fmt.Errorf("se superó el máximo de reintentos esperando el ACK para %d", expectedAckNum)
}
