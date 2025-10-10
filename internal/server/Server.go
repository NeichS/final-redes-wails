package server

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
)

// Server struct para gestionar el estado del servidor de forma segura
type Server struct {
	ctx         context.Context
	listener    net.Listener
	mu          sync.Mutex // Mutex para proteger el acceso concurrente
	isListening bool
}

func (s *Server) StartContext(ctx context.Context) {
	s.ctx = ctx
}

// ReceiveFileHandler inicia el servidor en una nueva goroutine para no bloquear la UI
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

	// Ejecutamos el bucle de aceptación en una goroutine para no bloquear la llamada
	go s.acceptLoop()

	return "Servidor iniciado en el puerto 8080", nil
}

// StopServerHandler detiene el servidor de forma segura
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
		go handleConnection(conn)
	}
}