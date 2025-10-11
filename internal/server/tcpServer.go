// internal/server/tcpServer.go

package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log"
	"net"
	"os"
)

func handleConnection(ctx context.Context, conn net.Conn) {
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

			for i := range reps {
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

					fmt.Fprintf(conn, "Segment %d received", i)
				}
			}

			newFile.Close()
			log.Printf("File %s received successfully.", fileName)
			fmt.Fprintf(conn, "File %s received successfully", fileName)
			runtime.EventsEmit(ctx, "reception-finished", fmt.Sprintf("¡%s recibido con éxito!", fileName))
		} else {
			log.Println("Invalid header received.")
		}
	}
}
