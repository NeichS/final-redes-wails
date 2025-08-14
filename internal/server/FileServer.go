package server

import (
	"context"
	"log"
)

type FileServer struct {
	ctx context.Context
}

type FileServerInfo struct {
	Address string
	TCP     bool
	Paths   []string
} 

func (fs *FileServer) StartContext(ctx context.Context) {
	fs.ctx = ctx
}

func (fs *FileServer) SendFile(fi FileServerInfo) (string, error) {
	log.Printf("Sending file to %s using TCP: %t with paths: %v", fi.Address, fi.TCP, fi.Paths)
	return "Server started", nil
}