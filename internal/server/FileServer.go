package server

import (
	"context"
	"log"
)

type FileServer struct {
	ctx context.Context
}

func (fs *FileServer) StartContext(ctx context.Context) {
	fs.ctx = ctx
}

func (fs *FileServer) SendFile(addr string, tcp bool, paths []string) (string, error) {
	log.Printf("Sending file to %s using TCP: %t with paths: %v", addr, tcp, paths)
	return "Server started", nil
}