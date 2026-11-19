package internal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeVisitorIgnoresLocalTypes(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "service.go", `package service
type Store interface { Read() error }
func check(x any) bool {
	type Store interface { Write() error }
	_, ok := x.(Store)
	return ok
}
`, 0)
	require.NoError(t, err)

	visitor := NewNodeVisitor(t.Context())
	ast.Walk(visitor, file)
	require.Len(t, visitor.DeclaredInterfaces(), 1)
	require.Equal(t, file.Decls[0], visitor.DeclaredInterfaces()[0].genDecl)
}
