// internal/client/tcpClient.go

package server

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath" // <--- Importa el paquete path/filepath

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
		log.Printf("Error dialing: %v", err)
		return err
	}
	defer conn.Close() // Asegura que la conexión se cierre al final

	// Llama a la función que ahora puede enviar múltiples archivos
	err = sendFiles(filePaths, conn)
	if err != nil {
		log.Printf("Error sending files: %v", err)
		return err
	}

	log.Println("All files sent successfully.")
	return nil
}

// Renombrada a sendFiles y ahora itera sobre los paths
func sendFiles(filePaths []string, conn *net.TCPConn) error {
	for _, path := range filePaths {
		err := sendSingleFile(path, conn)
		if err != nil {
			// Si hay un error con un archivo, lo reportamos y paramos
			return fmt.Errorf("failed to send file %s: %w", path, err)
		}
	}
	return nil
}

// Nueva función que encapsula la lógica para un solo archivo
func sendSingleFile(filePath string, conn *net.TCPConn) error {
	file, err := os.Open(filePath) // Usar os.Open para lectura
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return err
	}
	defer file.Close()

	baseName := filepath.Base(filePath)
	header := shared.NewMetadata(file, baseName)

	dataBuffer := make([]byte, 1014)

	headerBuffer := []byte{1}

	temp := make([]byte, 4)
	received := make([]byte, 1024)

	// Header: Número de segmentos
	binary.BigEndian.PutUint32(temp, header.Reps())
	headerBuffer = append(headerBuffer, temp...)

	// Header: Longitud del nombre
	binary.BigEndian.PutUint32(temp, uint32(len(header.Name())))
	headerBuffer = append(headerBuffer, temp...)

	// Header: Nombre
	headerBuffer = append(headerBuffer, []byte(header.Name())...)
	headerBuffer = append(headerBuffer, 0) // End of header

	_, err = conn.Write(headerBuffer)
	if err != nil {
		return err
	}

	// Esperar confirmación del header del servidor
	_, err = conn.Read(received)
	if err != nil {
		return err
	}
	fmt.Println(string(received))

	// Enviar segmentos
	for i := 0; i < int(header.Reps()); i++ {
		n, err := file.Read(dataBuffer)
		if err != nil && err != io.EOF {
			return err
		}

		segmentBuffer := []byte{0} // Start of segment

		// Segment number
		binary.BigEndian.PutUint32(temp, uint32(i))
		segmentBuffer = append(segmentBuffer, temp...)

		// Length of data
		binary.BigEndian.PutUint32(temp, uint32(n))
		segmentBuffer = append(segmentBuffer, temp...)

		// Data
		segmentBuffer = append(segmentBuffer, dataBuffer[:n]...)
		segmentBuffer = append(segmentBuffer, 1) // End of segment

		_, err = conn.Write(segmentBuffer)
		if err != nil {
			return err
		}

		// Esperar confirmación del segmento
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