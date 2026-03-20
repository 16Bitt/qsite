package qsite

import (
	"path"
	"strings"
)

type paths struct {
	root string
}

// Paths returns a path helper struct.
func (s *Server) Paths() paths {
	return paths{root: path.Clean(s.root)}
}

// BaseTemplateFSPath returns the relative path to the base go template that
// will be applied when rendering content.
func (p paths) BaseTemplateFSPath() string {
	return path.Join(p.root, "base.html.tmpl")
}

// StaticFSPath returns the relative path to the static asset directory.
func (p paths) StaticFSPath() string {
	return path.Join(p.root, "static")
}

// PagesFSPath returns the relative path to the pages directory containing the
// markdown content of the site.
func (p paths) PagesFSPath() string {
	return path.Join(p.root, "pages")
}

// FaviconFSPath is the filesystem path to favicon.ico.
func (p paths) FaviconFSPath() string {
	return path.Join(p.StaticFSPath(), "favicon.ico")
}

// DocWebPath returns the HTTP path of the given TreeNode's document.
func (p paths) DocWebPath(tn *TreeNode) string {
	return strings.TrimSuffix(p.subtractRoot(tn.fsPath), MarkdownExtension)
}

func (p paths) subtractRoot(sub string) string {
	// TODO: This is terrible, but the path package is frustratingly limited...
	child := strings.TrimPrefix(path.Clean(sub), p.PagesFSPath())
	return strings.TrimPrefix(child, "/")
}
