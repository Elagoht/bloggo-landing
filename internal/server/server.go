package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bloggo-landing/internal/webhook"
)

type Server struct {
	port            string
	prerenderedDir  string
	staticDir       string
	webhookHandler  *webhook.Handler
}

func New(port, prerenderedDir, staticDir string, webhookHandler *webhook.Handler) *Server {
	return &Server{
		port:           port,
		prerenderedDir: prerenderedDir,
		staticDir:      staticDir,
		webhookHandler: webhookHandler,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Webhook must be registered before the catch-all
	mux.HandleFunc("POST /webhook", s.webhookHandler.HandleWebhook)

	mux.HandleFunc("/static/", s.serveStatic)

	mux.HandleFunc("/", s.servePrerendered)

	log.Printf("Starting server on port %s", s.port)
	return http.ListenAndServe(":"+s.port, s.withLogging(mux))
}

func (s *Server) servePrerendered(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Skip webhook endpoint
	if path == "/webhook" {
		http.NotFound(w, r)
		return
	}

	if path == "/" {
		path = "/index.html"
	} else if !strings.HasSuffix(path, ".html") {
		path = path + ".html"
	}

	filePath := filepath.Join(s.prerenderedDir, path)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	http.ServeFile(w, r, filePath)
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	filePath := filepath.Join(s.staticDir, path)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")

	http.ServeFile(w, r, filePath)
}

func (s *Server) withLogging(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}

func (s *Server) Health() error {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/", s.port))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
