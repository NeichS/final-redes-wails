package server

import (
	"context"
)


type Server struct {
	ctx context.Context
}


func (c *Server) StartContext(ctx context.Context) {
	c.ctx = ctx
}


func (c *Server) ReceiveFileHandler() (string, error) {
	
	err := startTCPServer()

	if err != nil {
		return "", err
	}
	return "File received", nil
}
