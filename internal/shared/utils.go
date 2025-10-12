// internal/shared/utils.go

package shared

import (
	"log"
	"os"
)

type MetaData struct {
	name     string
	fileSize int64
	reps     uint32
	Checksum string
}

// Modificamos NewMetadata para aceptar el nombre del archivo por separado
func NewMetadata(file *os.File, baseName, checksum string) MetaData {
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	size := fileInfo.Size()

	header := MetaData{
		name:     baseName,
		fileSize: size,
		reps:     uint32(size/1014) + 1,
		Checksum: checksum,
	}

	return header
}


func (m *MetaData) Name() string {
	return m.name
}

func (m *MetaData) FileSize() int64 {
	return m.fileSize
}
func (m *MetaData) Reps() uint32 {
	return m.reps
}

func (m *MetaData) GetChecksum() string {
	return m.Checksum
}
