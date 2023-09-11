package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	ln     net.Listener
	server *http.Server

	router chi.Router

	Addr   string
	Domain string
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
	}

	s.router = chi.NewRouter()

	// No need to make the method part of Server.
	s.router.HandleFunc("/", handleHello)

	s.server.Handler = s.router

	return s
}

func (s *Server) Open() (err error) {
	if s.Domain != "" {
		s.ln = autocert.NewListener(s.Domain)
	} else {
		s.ln, err = net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
	}

	go s.server.Serve(s.ln)
	return nil
}

func (s *Server) Close() error {
	ctx, _ := context.WithTimeoutCause(context.Background(), time.Second*5, fmt.Errorf("shutdown: time exceeded"))
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}
	return "http"
}

func (s *Server) UseTLS() bool {
	return s.Domain != ""
}

func (s *Server) Port() int {
	if s.ln == nil {
		return -1
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	domain := "localhost"
	if s.Domain != "" {
		domain = s.Domain
	}

	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}

	return fmt.Sprintf("%s://%s:%d", scheme, domain, port)
}

// ListenAndServeTLSRedirect runs an HTTP server on port 80 to redirect users
// to the TLS-enabled port 443 server.
func ListenAndServeTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}
