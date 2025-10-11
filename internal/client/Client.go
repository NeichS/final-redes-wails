package server

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Client struct {
	ctx context.Context
}

type FileSenderInfo struct {
	Address string
	Port    string
	TCP     bool
	Paths   []string
}

func (c *Client) StartContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Client) SendFileHandler(fi FileSenderInfo) (string, error) {
	log.Printf("Sending file to %s using TCP: %t with paths: %v", fi.Address, fi.TCP, fi.Paths)

	if fi.TCP {
		err := startTCPClient(c.ctx, fi.Address, fi.Port, fi.Paths)

		if err != nil {
			log.Printf("Error starting TCP server: %v", err)
			runtime.EventsEmit(c.ctx, "client-error", "conexión rechazada")
			return "", err
		}
	}

	return "Server started", nil
}
