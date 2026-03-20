package qsite

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// MountDocPaths recursively iterates over the contents of a given TreeNode and
// "mounts" (creates an HTTP handler on the provided mux) the given directory.
func (s *Server) MountDocPaths(node *TreeNode, mux *http.ServeMux) error {
	logger := s.logger.With("stage", "mount")

	if node.Type == TreeNodeTypePage {
		logger.Debug("mounting doc", "path", node.fsPath)
		handler, err := s.docHandler(node)
		if err != nil {
			return err
		}

		mux.HandleFunc(s.Action(node), handler)
	} else {
		logger.Debug("recursing children", "path", node.fsPath)
		for _, child := range node.Children {
			err := s.MountDocPaths(child, mux)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) docHandler(tn *TreeNode) (http.HandlerFunc, error) {
	html, err := renderMarkdown(tn)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	input := TemplateInput{
		DocumentPath: s.Paths().DocWebPath(tn),
		Content:      template.HTML(html),
	}

	err = s.baseTemplate.Execute(buf, input)
	if err != nil {
		return nil, err
	}
	data := buf.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("serving document", "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Content-Type", "text/html")
		// TODO: Add separate TTL
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", s.staticTTL))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}, nil
}
