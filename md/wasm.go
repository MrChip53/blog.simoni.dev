package md

import (
	"github.com/gomarkdown/markdown/ast"
)

type WasmLoader struct {
	ast.Leaf
	Type    string // "go", "cpp", "rust", etc.
	WasmURL string
}
