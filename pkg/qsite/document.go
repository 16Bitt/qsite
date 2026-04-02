package qsite

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// renderMarkdown processes the document and returns the title (the first H1
// element's content), a string containing the processed HTML, and an error.
func renderMarkdown(tn *TreeNode) (string, string, error) {
	file, err := os.Open(tn.fsPath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	title := extractTitle(file)
	// Rather than opening the file twice, just rewind to the start after the
	// first scanning operation.
	file.Seek(0, io.SeekStart)

	raw, err := io.ReadAll(file)
	if err != nil {
		return title, "", err
	}

	return title, mdToHTML(raw), nil
}

func mdToHTML(raw []byte) string {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes | parser.SuperSubscript | parser.Attributes)
	doc := p.Parse(raw)
	renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})

	return string(markdown.Render(doc, renderer))
}

func extractTitle(file io.Reader) string {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	return ""
}
