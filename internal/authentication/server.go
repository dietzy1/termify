package authentication

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dietzy1/termify/internal/config"
	"github.com/zmb3/spotify/v2"
)

//go:embed login_complete.html
//go:embed login_failed.html
var content embed.FS

type server struct {
	config  *config.Config
	server  *http.Server
	service *service
}

func NewServer(c *config.Config, s *service) *server {

	mux := http.NewServeMux()
	srv := &server{
		config: c,
		server: &http.Server{
			Addr:    c.GetPort(),
			Handler: mux,
		},
		service: s,
	}

	mux.HandleFunc("/callback", srv.callbackHandler)
	mux.HandleFunc("/health", srv.healthHandler)

	return srv
}

func (s *server) ListenAndServe(ctx context.Context) error {
	errChan := make(chan error, 1)
	log.Println("Starting authentication service")
	go func() {
		log.Printf("Starting callback server on %s", s.config.GetPort())
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Shutdown()
	}
}

func (s *server) Shutdown() error {
	if s.server != nil {
		log.Println("Stopping callback server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
	return nil
}

func (s *server) callbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Callback received, processing authentication completion")
	verifierParams := s.service.verifier.params()

	tok, err := s.service.authenticator.Token(r.Context(), s.service.state, r, verifierParams...)
	if err != nil {
		log.Printf("Token error: %v", err)

		htmlContent, err := content.ReadFile("login_failed.html")
		if err != nil {
			log.Printf("Failed to load completion page: %v", err)
			http.Error(w, "Failed to load completion page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Writer.Write(w, htmlContent)
		return
	}
	log.Printf("Token successfully obtained")

	if st := r.FormValue("state"); st != s.service.state {
		log.Printf("State mismatch: expected %s, got %s", s.service.state, st)

		htmlContent, err := content.ReadFile("login_failed.html")
		if err != nil {
			log.Printf("Failed to load completion page: %v", err)
			http.Error(w, "Failed to load completion page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Writer.Write(w, htmlContent)
		return
	}

	log.Printf("State verification successful")

	client := spotify.New(s.service.authenticator.Client(r.Context(), tok))
	s.service.program.Send(
		LoginClientMsg{
			Client: client,
		},
	)

	htmlContent, err := content.ReadFile("login_complete.html")
	if err != nil {
		log.Printf("Failed to load completion page: %v", err)
		http.Error(w, "Failed to load completion page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Writer.Write(w, htmlContent)
	log.Println("Callback processing complete, HTML response sent")

	if err := s.service.saveToken(tok); err != nil {
		log.Printf("Failed to save token: %v", err)
		return
	}
	log.Printf("Token successfully saved")
	if err := s.service.saveClientID(s.service.clientID); err != nil {
		log.Printf("Failed to save client ID: %v", err)
		return
	}

}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
