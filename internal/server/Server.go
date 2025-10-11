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
		return "", errors.New("el servidor ya está escuchando")
	}
	
    listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		s.mu.Unlock()
		log.Printf("Error listening on port 8080: %v", err)
		return "", err
	}
	
    s.listener = listener
	s.isListening = true
	s.mu.Unlock()

	log.Println("Server listening on port 8080")

	go s.acceptLoop()

	return "Servidor iniciado en el puerto 8080", nil
}

func (s *Server) StopServerHandler() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isListening && s.listener != nil {
		log.Println("Deteniendo el servidor...")
		s.isListening = false
		s.listener.Close() // Esto hará que Accept() devuelva un error
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