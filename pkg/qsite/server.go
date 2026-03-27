package qsite

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// Server is an instance of a qsite server.
type Server struct {
	addr         string
	root         string
	staticTTL    int
	logger       *slog.Logger
	baseTemplate *template.Template
	env          string
}

// TemplateInput is the data available to the base template when rendering the
// site.
type TemplateInput struct {
	DocumentPath string
	DocumentName string
	Env          string
	Content      template.HTML
}

func NewServer(addr, dir string, staticTTL int, logLevel, env string) *Server {
	opts := &slog.HandlerOptions{Level: toLoglevel(logLevel)}
	handler := slog.NewTextHandler(os.Stdout, opts)

	return &Server{
		addr:      addr,
		root:      dir,
		staticTTL: staticTTL,
		logger:    slog.New(handler).With("layer", "server").With("env", env),
		env:       env,
	}
}

// Listen loads the content and blocks while serving the content.
func (s *Server) Listen() error {
	tree, err := s.PageTree()
	if err != nil {
		return err
	}

	t, err := template.New("base.html.tmpl").Funcs(s.templateHelpers()).ParseFiles(s.Paths().BaseTemplateFSPath())
	if err != nil {
		return err
	}
	s.baseTemplate = t.Funcs(s.templateHelpers())

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

func (s *Server) templateHelpers() map[string]any {
	return map[string]any{
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"join": func(values ...string) string {
			return strings.Join(values, "")
		},
		"isDev": func() bool {
			return s.env == "dev"
		},
		"raw": func(value string) template.HTML {
			return template.HTML(value)
		},
	}
}
