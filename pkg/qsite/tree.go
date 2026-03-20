package qsite

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const MarkdownExtension = ".md"

type TreeNodeType int

const (
	TreeNodeTypePage = iota
	TreeNodeTypeDir
)

// TreeNode represents a node within the tree of documents for the project.
type TreeNode struct {
	Type     TreeNodeType
	Children []*TreeNode

	fsPath string
}

func NewPageNode(path string) *TreeNode {
	return &TreeNode{
		Type:   TreeNodeTypePage,
		fsPath: path,
	}
}

func NewDirNode(path string) *TreeNode {
	return &TreeNode{
		Type:     TreeNodeTypeDir,
		fsPath:   path,
		Children: make([]*TreeNode, 0, 0),
	}
}

func (tn *TreeNode) DocumentName() string {
	return path.Base(strings.TrimSuffix(tn.fsPath, MarkdownExtension))
}

// PageTree recursively builds the document tree for the site.
func (s *Server) PageTree() (*TreeNode, error) {
	root := NewDirNode(s.Paths().PagesFSPath())
	err := s.buildTree(root)
	return root, err
}

// Action returns the full HTTP verb+path text for the given document node.
func (s *Server) Action(tn *TreeNode) string {
	return fmt.Sprintf("GET /%s", s.Paths().DocWebPath(tn))
}

func (s *Server) buildTree(root *TreeNode) error {
	entries, err := os.ReadDir(root.fsPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if isHidden(entry) {
			continue
		}

		if entry.IsDir() {
			child := root.childDirNode(entry.Name())
			s.buildTree(child)
		} else if isDocumentFile(entry) {
			root.childPageNode(entry.Name())
		}
	}

	return nil
}

func isDocumentFile(entry os.DirEntry) bool {
	return entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), MarkdownExtension)
}

func isHidden(entry os.DirEntry) bool {
	return strings.HasPrefix(entry.Name(), ".")
}

func (tn *TreeNode) childPageNode(name string) *TreeNode {
	if tn.Type != TreeNodeTypeDir {
		panic(fmt.Sprintf("tried to append child to non-directory path %s", tn.fsPath))
	}

	child := NewPageNode(path.Join(tn.fsPath, name))
	tn.Children = append(tn.Children, child)

	return child
}

func (tn *TreeNode) childDirNode(name string) *TreeNode {
	if tn.Type != TreeNodeTypeDir {
		panic(fmt.Sprintf("tried to append child to non-directory path %s", tn.fsPath))
	}

	child := NewDirNode(path.Join(tn.fsPath, name))
	tn.Children = append(tn.Children, child)

	return child
}
