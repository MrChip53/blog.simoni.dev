package server

import (
	"bytes"
	"fmt"
	"time"

	"blog.simoni.dev/md"
	"blog.simoni.dev/models"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func formatAsDateTime(t time.Time) string {
	year, month, day := t.Date()
	dateString := fmt.Sprintf("%d/%02d/%02d", year, month, day)

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		fmt.Println("Error loading time zone:", err)
		return dateString
	}

	chicagoTime := t.In(loc)

	timeFormat := "01/02/2006 3:04 PM"

	timeString := chicagoTime.Format(timeFormat)

	return timeString
}

func parseMarkdown(bytes []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	p.RegisterInline(':', parseWasmLoader)
	doc := p.Parse(bytes)

	renderer := md.NewRenderer()

	return markdown.Render(doc, renderer)
}

// Syntax: ::wasm[type](url)
// Example: ::wasm[game](https://example.com/game.wasm)
func parseWasmLoader(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	// Check for ::wasm prefix
	if !bytes.HasPrefix(data[offset:], []byte("::wasm[")) {
		return 0, nil
	}

	i := offset + 7 // len("::wasm[")

	// Extract type
	typeStart := i
	for i < len(data) && data[i] != ']' {
		i++
	}
	if i >= len(data) {
		return 0, nil
	}
	wasmType := string(data[typeStart:i])
	i++ // skip ']'

	// Check for (
	if i >= len(data) || data[i] != '(' {
		return 0, nil
	}
	i++ // skip '('

	// Extract URL
	urlStart := i
	for i < len(data) && data[i] != ')' {
		i++
	}
	if i >= len(data) {
		return 0, nil
	}
	wasmURL := string(data[urlStart:i])
	i++ // skip ')'

	node := &md.WasmLoader{
		Type:    wasmType,
		WasmURL: wasmURL,
	}

	return i - offset, node
}

func getSlug(post models.BlogPost) string {
	return fmt.Sprintf("/post/%02d/%02d/%d/%s", post.CreatedAt.Month(), post.CreatedAt.Day(), post.CreatedAt.Year(), post.Slug)
}

func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}

	return s
}
