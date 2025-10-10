package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("Accepted new connection.")

	// Buffer para leer los datos entrantes
	buffer := make([]byte, 2048) // Aumentamos el buffer por si los nombres son largos

	// Leer el header inicial
	n, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			log.Printf("Error reading from connection: %v", err)
		}
		return
	}

	// [1 (1 byte), reps (4 bytes), nameLen (4 bytes), name (n bytes), 0 (1 byte)]
	if n > 10 && buffer[0] == 1 && buffer[n-1] == 0 {
		reps := binary.BigEndian.Uint32(buffer[1:5])
		nameLen := binary.BigEndian.Uint32(buffer[5:9])

		if (9 + nameLen) > uint32(n-1) {
			log.Println("Invalid header: name length is too long.")
			return
		}
		fileName := string(buffer[9 : 9+nameLen])

		log.Printf("Receiving file: %s, Segments: %d", fileName, reps)

		// Confirmación al cliente
		conn.Write([]byte("Header received"))

		// Crear el archivo para escribir los datos
		newFile, err := os.Create(fileName)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			return
		}
		defer newFile.Close()

		// Recibir los segmentos del archivo
		for i := uint32(0); i < reps; i++ {
			n, err = conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading segment: %v", err)
				}
				break
			}

			// [0 (1 byte), segNum (4 bytes), dataLen (4 bytes), data (n bytes), 1 (1 byte)]
			if n > 10 && buffer[0] == 0 && buffer[n-1] == 1 {
				// segNum := binary.BigEndian.Uint32(buffer[1:5])
				dataLen := binary.BigEndian.Uint32(buffer[5:9])

				if (9 + dataLen) > uint32(n-1) {
					log.Println("Invalid segment: data length is too long.")
					return
				}
				data := buffer[9 : 9+dataLen]


				_, err := newFile.Write(data)
				if err != nil {
					log.Printf("Error writing to file: %v", err)
					return
				}

				// Confirmación de segmento recibido
				conn.Write([]byte(fmt.Sprintf("Segment %d received", i)))
			}
		}

		log.Printf("File %s received successfully.", fileName)
		conn.Write([]byte("File received successfully"))

	} else {
		log.Println("Invalid header received.")
	}
}