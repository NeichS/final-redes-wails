package server

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
)

type Server struct {
	ctx         context.Context
	listener    net.Listener
	mu          sync.Mutex
	isListening bool
}

func (s *Server) StartContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Server) ReceiveFileHandler() (string, error) {
	s.mu.Lock()
	if s.isListening {
		s.mu.Unlock()
		return "", errors.New("los servidores ya están escuchando")
	}
	s.isListening = true
	s.mu.Unlock()

	tcpListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		s.StopServerHandler()
		return "", err
	}
	s.listener = tcpListener
	log.Println("Servidor TCP escuchando en :8080")
	go s.acceptLoop()

	go s.startUDPServer()

	return "Servidores TCP y UDP iniciados en puerto 8080", nil
}

func (s *Server) StopServerHandler() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isListening {
		log.Println("Deteniendo servidores...")
		s.isListening = false
		if s.listener != nil {
			s.listener.Close()
		}
	}
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			if !s.isListening {
				s.mu.Unlock()
				log.Println("Servidor detenido correctamente.")
				return // Salimos del bucle y de la goroutine
			}
			s.mu.Unlock()
			log.Printf("Error al aceptar la conexión: %v", err)
			continue
		}
		go handleConnection(s.ctx, conn)
	}
}