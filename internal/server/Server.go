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
	s.isListening = true // Marcar como escuchando
	s.mu.Unlock()

	// Iniciar listener TCP en una goroutine
	tcpListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		s.StopServerHandler()
		return "", err
	}
	s.listener = tcpListener // Guardar para poder cerrarlo después
	log.Println("Servidor TCP escuchando en :8080")
	go s.acceptLoop()

	// Iniciar listener UDP en otra goroutine
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

// Bucle para aceptar conexiones
func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Verificamos si el error es porque cerramos la conexión a propósito
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
		// Maneja cada conexión en su propia goroutine
		go handleConnection(s.ctx, conn)
	}
}