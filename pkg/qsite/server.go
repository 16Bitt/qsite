package qsite

import (
	"html/template"
	"log/slog"
	"net/http"
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

func NewServer(addr, dir string, staticTTL int) *Server {
	return &Server{
		addr:      addr,
		root:      dir,
		staticTTL: staticTTL,
		logger:    slog.Default().With("layer", "server"),
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

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
	})

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, s.Paths().FaviconFSPath())
	})

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	return server.ListenAndServe()
}
