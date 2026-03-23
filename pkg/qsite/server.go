package qsite

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
)

// Server is an instance of a qsite server.
type Server struct {
	addr         string
	root         string
	staticTTL    int
	logger       *slog.Logger
	baseTemplate *template.Template
}

// TemplateInput is the data available to the base template when rendering the
// site.
type TemplateInput struct {
	DocumentPath string
	Content      template.HTML
}

func NewServer(addr, dir string, staticTTL int, logLevel string) *Server {
	opts := &slog.HandlerOptions{Level: toLoglevel(logLevel)}
	handler := slog.NewTextHandler(os.Stdout, opts)

	return &Server{
		addr:      addr,
		root:      dir,
		staticTTL: staticTTL,
		logger:    slog.New(handler).With("layer", "server"),
	}
}

// Listen loads the content and blocks while serving the content.
func (s *Server) Listen() error {
	tree, err := s.PageTree()
	if err != nil {
		return err
	}

	t, err := template.ParseFiles(s.Paths().BaseTemplateFSPath())
	if err != nil {
		return err
	}
	s.baseTemplate = t

	mux := http.NewServeMux()
	err = s.MountDocPaths(tree, mux)
	if err != nil {
		return err
	}

	fs := http.FileServer(http.Dir(s.Paths().StaticFSPath()))
	mux.Handle("/static/", WrapCache(http.StripPrefix("/static", fs), s.staticTTL))

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, s.Paths().FaviconFSPath())
	})

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	return server.ListenAndServe()
}

func toLoglevel(name string) slog.Level {
	switch name {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
