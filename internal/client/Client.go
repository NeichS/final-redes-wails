package server

import (
	"context"
	"log"
	"sync"
)

type Client struct {
	ctx        context.Context
	downtime   bool
	downtimeMu sync.RWMutex
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

func (c *Client) ToggleDowntime(active bool) {
	c.downtimeMu.Lock()
	defer c.downtimeMu.Unlock()
	c.downtime = active
	if active {
		log.Println("Downtime started")
	} else {
		log.Println("Downtime ended")
	}
}

func (c *Client) IsDowntime() bool {
	c.downtimeMu.RLock()
	defer c.downtimeMu.RUnlock()
	return c.downtime
}

func (c *Client) SendFileHandler(fi FileSenderInfo) (string, error) {
	protocol := "UDP"
	if fi.TCP {
		protocol = "TCP"
	}
	log.Printf("Sending file to %s using %s, with paths: %v", fi.Address, protocol, fi.Paths)

	if fi.TCP {
		err := startTCPClient(c.ctx, fi.Address, fi.Port, fi.Paths, c)

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
