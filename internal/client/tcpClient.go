// internal/client/tcpClient.go

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
	"path/filepath"
	"time"

	"github.com/NeichS/final-redes-wails/internal/shared"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func startTCPClient(ctx context.Context, addr, port string, filePaths []string) error {
	tcpServer, err := net.ResolveTCPAddr("tcp", addr+":"+port)
	if err != nil {
		log.Printf("Error resolving TCP address: %v", err)
		runtime.EventsEmit(ctx, "client-error", "dirección IP inválida")
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("No se pudo conectar: %v", err))
		log.Printf("Error dialing: %v", err)
		return err
	}
	defer conn.Close()

	err = sendFiles(ctx, filePaths, conn)
	if err != nil {
		log.Printf("Error sending files: %v", err)
		runtime.EventsEmit(ctx, "client-error", fmt.Sprintf("Error durante el envío: %v", err))
		return err
	}

	runtime.EventsEmit(ctx, "reception-finished", "¡Todos los archivos enviados con éxito!")
	return nil
}

// Renombrada a sendFiles y ahora itera sobre los paths
func sendFiles(ctx context.Context, filePaths []string, conn *net.TCPConn) error {
	totalFiles := len(filePaths)
	for i, path := range filePaths {
		runtime.EventsEmit(ctx, "sending-file-start", map[string]interface{}{
			"fileName":    filepath.Base(path),
			"currentFile": i + 1,
			"totalFiles":  totalFiles,
		})
		time.Sleep(100 * time.Millisecond)
		err := sendSingleFile(ctx, path, conn)
		if err != nil {
			// Si hay un error con un archivo, lo reportamos y paramos
			return fmt.Errorf("failed to send file %s: %w", path, err)
		}
	}
	return nil
}

func sendSingleFile(ctx context.Context, filePath string, conn *net.TCPConn) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Printf("Error calculating checksum: %v", err)
		return err
	}
	checksum := hex.EncodeToString(hash.Sum(nil))
	file.Seek(0, 0)

	baseName := filepath.Base(filePath)
	header := shared.NewMetadata(file, baseName, checksum)

	dataBuffer := make([]byte, 1014)

	headerBuffer := []byte{1}

	temp := make([]byte, 4)
	received := make([]byte, 1024)

	binary.BigEndian.PutUint32(temp, header.Reps())
	headerBuffer = append(headerBuffer, temp...)

	binary.BigEndian.PutUint32(temp, uint32(len(header.Name())))
	headerBuffer = append(headerBuffer, temp...)

	binary.BigEndian.PutUint32(temp, uint32(len(header.GetChecksum())))
	headerBuffer = append(headerBuffer, temp...)

	headerBuffer = append(headerBuffer, []byte(header.Name())...)
	headerBuffer = append(headerBuffer, []byte(header.GetChecksum())...)
	headerBuffer = append(headerBuffer, 0)

	_, err = conn.Write(headerBuffer)
	if err != nil {
		return err
	}

	_, err = conn.Read(received)
	if err != nil {
		return err
	}
	fmt.Println(string(received))

	for i := 0; i < int(header.Reps()); i++ {
		n, err := file.Read(dataBuffer)
		if err != nil && err != io.EOF {
			return err
		}

		segmentBuffer := []byte{0}

		binary.BigEndian.PutUint32(temp, uint32(i))
		segmentBuffer = append(segmentBuffer, temp...)

		binary.BigEndian.PutUint32(temp, uint32(n))
		segmentBuffer = append(segmentBuffer, temp...)

		segmentBuffer = append(segmentBuffer, dataBuffer[:n]...)
		segmentBuffer = append(segmentBuffer, 1) // End of segment

		_, err = conn.Write(segmentBuffer)
		if err != nil {
			return err
		}

		runtime.EventsEmit(ctx, "sending-file-progress", map[string]interface{}{
			"sent":  i + 1,
			"total": header.Reps(),
		})

		_, err = conn.Read(received)
		if err != nil {
			if err == io.EOF {
				return errors.New("connection closed by server prematurely")
			}
			return err
		}
		fmt.Println(string(received))
	}
	return nil
}
