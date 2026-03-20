package qsite

import (
	"io"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func renderMarkdown(tn *TreeNode) (string, error) {
	file, err := os.Open(tn.fsPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return mdToHTML(raw), nil
}

func mdToHTML(raw []byte) string {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes)
	doc := p.Parse(raw)
	renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})

	return string(markdown.Render(doc, renderer))
}
