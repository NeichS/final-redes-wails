package server

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
    "github.com/NeichS/final-redes-wails/internal/shared"
)



func startTCPClient(addr, port string, filePaths []string) error {

	tcpServer, err := net.ResolveTCPAddr("tcp", addr+":"+port)

	if err != nil {
		log.Printf("Error resolving TCP address: %v", err)
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		log.Fatal(err)
	}

	sendFile(filePaths, conn)

	received := make([]byte, 1024)

	_, err = conn.Read(received)

	if err != nil {
		log.Fatal(err)
	}

	println(string(received))

	return nil
}

func sendFile(filePaths []string, conn *net.TCPConn) {
	 file, err := os.OpenFile(filePaths[0], os.O_RDONLY, 0755)

    if err != nil {
        log.Fatal(err)
    }

    header := shared.NewMetadata(file)

    dataBuffer := make([]byte, 1014)

    // Start (all 1s) - 1 byte, reps - 4 bytes, lengthofname - 4 bytes, name - `lengthofname` bytes, End (all 0s) - 1 byte;
    headerBuffer := []byte{1}

    // Start (all 0s) - 1 byte, Segment number - 4 bytes, lengthofdata - 4 bytes, Data - `lengthofdata` bytes, End (all 1s) - 1 byte
    segmentBuffer := []byte{0}

    // Temporary buffer for uint32
    temp := make([]byte, 4)

    // Temporary buffer for responses received
    received := make([]byte, 100);

    for i := 0; i < int(header.Reps()); i++ {
        n, _ := file.ReadAt(dataBuffer, int64(i*1014))

        if i == 0 {
            // Number of segments
            binary.BigEndian.PutUint32(temp, header.Reps())
            headerBuffer = append(headerBuffer, temp...)

            // Length of name
            binary.BigEndian.PutUint32(temp, uint32(len(header.Name())))
            headerBuffer = append(headerBuffer, temp...)

            // Name
            headerBuffer = append(headerBuffer, []byte(header.Name())...)

            headerBuffer = append(headerBuffer, 0)

            _, err := conn.Write(headerBuffer)

            if err != nil {
                log.Fatal(err)
            }

            _, err = conn.Read(received)

            if err != nil {
                log.Fatal(err)
            }

            println(string(received))
        }

        // Segment number
        binary.BigEndian.PutUint32(temp, uint32(i))
        segmentBuffer = append(segmentBuffer, temp...);

        // Length of data
        binary.BigEndian.PutUint32(temp, uint32(n))
        segmentBuffer = append(segmentBuffer, temp...)

        // Data
        segmentBuffer = append(segmentBuffer, dataBuffer...)

        segmentBuffer = append(segmentBuffer, 1)

        _, err = conn.Write(segmentBuffer);

        if err != nil {
            log.Fatal(err);
        }

        _, err = conn.Read(received);
        fmt.Println(string(received));

        if err != nil {
            log.Fatal(err)
        }

        // Reset segment buffer
        segmentBuffer = []byte{0};
    }
}


