package shared

import (
	"log"
	"os"
)

type MetaData struct {
    name     string
    fileSize uint32
    reps     uint32
}

func NewMetadata(file *os.File) MetaData {

    fileInfo, err := file.Stat()

    if err != nil {
        log.Fatal(err)
    }

    size := fileInfo.Size()

    header := MetaData{
        name:     file.Name(),
        fileSize: uint32(size),
        reps:     uint32(size / 1014) + 1,
    }

    return header
}

func (m *MetaData) Name() string {
	return m.name
}

func (m *MetaData) FileSize() uint32 {
	return m.fileSize
}
func (m *MetaData) Reps() uint32 {
	return m.reps
}