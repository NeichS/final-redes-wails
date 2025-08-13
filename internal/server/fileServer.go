package server

import "context"

type FileServer struct {
	ctx context.Context
}

func (fs *FileServer) StartContext(ctx context.Context) {
	fs.ctx = ctx
}