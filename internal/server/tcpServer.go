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

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.Println("Accepted new connection, waiting for files...")

	for {
		msgType := make([]byte, 1)
		_, err := io.ReadFull(conn, msgType)
		if err != nil {
			if err == io.EOF {
				log.Println("Client closed connection cleanly.")
			} else {
				log.Printf("Error reading message type: %v", err)
			}
			return
		}

		if msgType[0] != 1 {
			log.Printf("Invalid message type received. Expected header (1), got (%d)", msgType[0])
			runtime.EventsEmit(ctx, "server-error", "Error de sincronización con el cliente.")
			return
		}

		headerFields := make([]byte, 12) // 4 bytes para reps + 4 para nameLen
		_, err = io.ReadFull(conn, headerFields)
		if err != nil {
			log.Printf("Error reading header fields: %v", err)
			return
		}
		reps := binary.BigEndian.Uint32(headerFields[0:4])
		nameLen := binary.BigEndian.Uint32(headerFields[4:8])
		checksumLen := binary.BigEndian.Uint32(headerFields[8:12])

		payloadAndEndByte := make([]byte, nameLen+checksumLen+1)
		_, err = io.ReadFull(conn, payloadAndEndByte)
		if err != nil {
			log.Printf("Error reading header payload: %v", err)
			return
		}

		if payloadAndEndByte[nameLen+checksumLen] != 0 {
			log.Println("Invalid header: missing end byte.")
			continue
		}
		fileName := string(payloadAndEndByte[:nameLen])
		receivedChecksum := string(payloadAndEndByte[nameLen : nameLen+checksumLen])

		log.Printf("Receiving file: %s, Segments: %d", fileName, reps)
		runtime.EventsEmit(ctx, "reception-started", fileName)
		conn.Write([]byte("Header received for " + fileName))

		if err := os.MkdirAll("./receive", 0755); err != nil {
			log.Printf("Error creating directory: %v", err)
			continue
		}

		newFile, err := os.Create("./receive/" + fileName)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			continue
		}

		for i := uint32(0); i < reps; i++ {
			segmentHeader := make([]byte, 9)
			_, err := io.ReadFull(conn, segmentHeader)
			if err != nil {
				log.Printf("Error reading segment header: %v", err)
				newFile.Close()
				return
			}

			if segmentHeader[0] != 0 {
				log.Println("Invalid segment: wrong type byte.")
				newFile.Close()
				return
			}

			dataLen := binary.BigEndian.Uint32(segmentHeader[5:9])
			dataBuffer := make([]byte, dataLen+1) // +1 para el byte final

			_, err = io.ReadFull(conn, dataBuffer)
			if err != nil {
				log.Printf("Error reading segment data: %v", err)
				newFile.Close()
				return
			}

			if dataBuffer[dataLen] != 1 {
				log.Println("Invalid segment: missing end byte.")
				newFile.Close()
				return
			}

			// Escribir en el archivo
			_, err = newFile.Write(dataBuffer[:dataLen])
			if err != nil {
				log.Printf("Error writing to file: %v", err)
				newFile.Close()
				return
			}
			// Enviar confirmación del segmento
			fmt.Fprintf(conn, "Segment %d received", i)
		}

		newFile.Close()
		log.Printf("File %s received successfully.", fileName)

		fileToVerify, err := os.Open("./receive/" + fileName)
		if err != nil {
			log.Printf("Could not open received file for verification: %v", err)
			continue
		}
		defer fileToVerify.Close()

		hash := md5.New()
		if _, err := io.Copy(hash, fileToVerify); err != nil {
			log.Printf("Error calculating checksum for received file: %v", err)
			continue
		}
		calculatedChecksum := hex.EncodeToString(hash.Sum(nil))

		if receivedChecksum == calculatedChecksum {
			log.Println("Checksums match! File is intact.")
			runtime.EventsEmit(ctx, "reception-finished", fmt.Sprintf("✅ ¡%s recibido y verificado con éxito!", fileName))
		} else {
			log.Println("CHECKSUM MISMATCH! File is corrupted.")
			runtime.EventsEmit(ctx, "server-error", fmt.Sprintf("❌ Error de checksum en %s. El archivo está corrupto.", fileName))
		}
	}
}
