package md

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"blog.simoni.dev/templates/components"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/google/uuid"
)

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func renderCode(w io.Writer, codeBlock *ast.CodeBlock, entering bool) {
	lang := string(codeBlock.Info)
	htmlHighlight(w, string(codeBlock.Literal), lang, "")
}

func renderWasmLoader(w io.Writer, wasm *WasmLoader) {
	//loaderId := uuid.New().String()

	//io.WriteString(w, fmt.Sprintf("<div id=\"wasm-%s\" data-type=\"%s\" data-wasm-url=\"%s\">", loaderId, wasm.Type, wasm.WasmURL))

	// use Ifram to wasm loader
	io.WriteString(w, fmt.Sprintf("<iframe src=\"/wasm/%s?url=%s\" onload=\"resizeIframe(this)\" style=\"width: 100%%;\"></iframe>", wasm.Type, wasm.WasmURL))
}

func renderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if code, ok := node.(*ast.CodeBlock); ok {
		b64Data := base64.StdEncoding.EncodeToString(code.Literal)
		io.WriteString(w, "<div class=\"code-block-wrapper\">")
		copyId := uuid.New().String()
		copyButton := components.CopyButton(b64Data, "copyBtn-"+copyId)
		copyButton.Render(context.TODO(), w)
		renderCode(w, code, entering)
		io.WriteString(w, "</div>")
		return ast.GoToNext, true
	}

	if wasm, ok := node.(*WasmLoader); ok && entering {
		renderWasmLoader(w, wasm)
		return ast.GoToNext, true
	}

	return ast.GoToNext, false
}

func NewRenderer() *mdhtml.Renderer {
	// init formatter and style
	if htmlFormatter == nil {
		htmlFormatter = html.New(html.TabWidth(4), html.WithClasses(true), html.WithLineNumbers(true))
		if htmlFormatter == nil {
			panic("couldn't create html formatter")
		}
	}
	if highlightStyle == nil {
		styleName := "monokai"
		highlightStyle = styles.Get(styleName)
		if highlightStyle == nil {
			panic("couldn't get highlight style")
		}
	}
	opts := mdhtml.RendererOptions{
		Flags:          mdhtml.CommonFlags | mdhtml.HrefTargetBlank,
		RenderNodeHook: renderHook,
	}
	return mdhtml.NewRenderer(opts)
}
