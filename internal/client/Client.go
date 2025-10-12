package server

import (
	"context"
	"log"

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
	protocol := "UDP"
	if fi.TCP {
		protocol = "TCP"
	}
	log.Printf("Sending file to %s using %s, with paths: %v", fi.Address, protocol, fi.Paths)

	if fi.TCP {
		err := startTCPClient(c.ctx, fi.Address, fi.Port, fi.Paths)

		if err != nil {
			log.Printf("Error starting TCP server: %v", err)
			return "", err
		}
	} else {
		err := startUDPClient(c.ctx, fi.Address, fi.Port, fi.Paths)
		if err != nil {
			log.Printf("Error starting UDP client: %v", err)
			return "", err
		}
	}

	return "Server started", nil
}
