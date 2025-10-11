// internal/server/tcpServer.go

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
	log.Println("Accepted new connection, waiting for files...")

	for {
		buffer := make([]byte, 2048)

		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Client closed the connection. Server is ready for a new one.")
			} else {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		if n > 10 && buffer[0] == 1 && buffer[n-1] == 0 {
			reps := binary.BigEndian.Uint32(buffer[1:5])
			nameLen := binary.BigEndian.Uint32(buffer[5:9])

			if (9 + nameLen) > uint32(n-1) {
				log.Println("Invalid header: name length is too long.")
				continue
			}
			fileName := string(buffer[9 : 9+nameLen])

			log.Printf("Receiving file: %s, Segments: %d", fileName, reps)

			conn.Write([]byte("Header received for " + fileName))

			newFile, err := os.Create(fileName)
			if err != nil {
				log.Printf("Error creating file: %v", err)
				continue
			}

			// Recibir los segmentos del archivo
			for i := uint32(0); i < reps; i++ {
				n, err = conn.Read(buffer)
				if err != nil {
					if err != io.EOF {
						log.Printf("Error reading segment: %v", err)
					}
					break
				}

				if n > 10 && buffer[0] == 0 && buffer[n-1] == 1 {
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

					conn.Write([]byte(fmt.Sprintf("Segment %d received", i)))
				}
			}

			newFile.Close()
			log.Printf("File %s received successfully.", fileName)
			conn.Write([]byte(fmt.Sprintf("File %s received successfully", fileName)))

		} else {
			log.Println("Invalid header received.")
		}
	}
}
