// The qsite package provides a basic HTTP server that serves markdown content
// and static content with a small handful of conventions to make authoring
// content easier.
//
// The expected directory structure is as follows:
//
//	/base.html.tmpl       - HTML template, used to render all content
//	/pages/**/*.md 	      - Directory containing all markdown content
//	/pages/index.md       - Will get rendered and served as /
//	/pages/404.md         - Will get rendered and served for 404 responses
//	/static/**/*          - Content will be served as-is, use for stylesheets, images, etc.
//	/static/favicon.ico   - Special cased: will get served under the root as /favicon.ico
//
// Markdown content is processed at boot time, so the server will need to be
// restarted to reflect changes. This simplifies routing significantly, and
// removes the need for any filesystem-to-HTTP conversions outside of the
// static content.
package qsite

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type ServerOptions struct {
	Addr           string // TCP listen address
	Root           string // Site root, must contain pages/, static/, and base.html.tmpl
	StaticTTL      int    // Max cache age for static content (applies to pages as well)
	Env            string // Environment name, can be used in the template to change behaviors.
	MetricsEnabled bool   // If true, expose /_metrics
	LogLevel       string // Logging verbosity, must be one of: debug, info, warn or error
}

// Server is an instance of a qsite server.
type Server struct {
	addr           string
	root           string
	staticTTL      int
	logger         *slog.Logger
	baseTemplate   *template.Template
	env            string
	metricsEnabled bool
	stats          *Stats
}

// TemplateInput is the data available to the base template when rendering the
// site.
type TemplateInput struct {
	Title        string // derived from the first h1 in the content
	DocumentPath string
	DocumentName string
	Env          string
	Content      template.HTML
}

func NewServer(opts ServerOptions) *Server {
	logOpts := &slog.HandlerOptions{Level: toLoglevel(opts.LogLevel)}
	handler := slog.NewTextHandler(os.Stdout, logOpts)

	return &Server{
		addr:           opts.Addr,
		root:           opts.Root,
		staticTTL:      opts.StaticTTL,
		logger:         slog.New(handler).With("layer", "server").With("env", opts.Env),
		env:            opts.Env,
		stats:          NewStats(),
		metricsEnabled: opts.MetricsEnabled,
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

	if s.metricsEnabled {
		mux.HandleFunc("GET /_metrics", s.stats.Handler())
	}

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
