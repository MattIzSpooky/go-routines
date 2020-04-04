package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"routine-server/spammer"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	http.Server
	logger           *log.Logger
	messageChan      chan string
	mux              *sync.Mutex
	requestsReceived int
}

func NewServer(address string, port string) *Server {
	var fullAddress string

	if runtime.GOOS != "windows" && address == "localhost" {
		fullAddress = fmt.Sprintf(":%s", port)
	} else {
		fullAddress = fmt.Sprintf("%s:%s", address, port)
	}

	server := &Server{
		Server: http.Server{
			Addr:         fullAddress,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
		logger:           log.New(os.Stdout, "Log: ", 0),
		messageChan:      make(chan string),
		mux:              &sync.Mutex{},
		requestsReceived: 0,
	}

	server.addHandlers()

	return server
}

func (s *Server) addHandlers() {
	mux := &http.ServeMux{}

	mux.Handle("/spam", s.incrementMiddleware(s.loggerMiddleware(http.HandlerFunc(s.spamHandler))))

	s.Handler = mux
}

func (s Server) spamHandler(w http.ResponseWriter, r *http.Request) {
	var spam spammer.Spam
	err := json.NewDecoder(r.Body).Decode(&spam)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.messageChan <- fmt.Sprintf("Name: %s, Email: %s", spam.Name, spam.Email)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println(fmt.Sprintf("Request #%d accepted from: %s", s.requestsReceived, r.RemoteAddr))

		next.ServeHTTP(w, r)
	})
}

func (s *Server) incrementMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mux.Lock()
		s.requestsReceived++
		s.mux.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (s Server) Listen() error {
	s.logger.Println(fmt.Sprintf("Running HTTP server on: http://%s", s.Addr))

	go func() {
		for message := range s.messageChan {
			s.logger.Println(fmt.Sprintf("Received: %s", message))
		}
	}()

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatalf("Could not listen on %s: %v\n", s.Addr, err)

		return err
	}

	return nil
}

func (s Server) GracefulShutdown() error {
	s.logger.Println("Server is shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.SetKeepAlivesEnabled(false)

	if err := s.Shutdown(ctx); err != nil {
		s.logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		return err
	}

	close(s.messageChan)
	s.logger.Println("Server has shutdown")

	return nil
}
