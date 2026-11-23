package internal

import (
	"context"
	"fmt"
	"go/ast"

	"github.com/rs/zerolog"
)

type declaredInterface struct {
	typeSpec *ast.TypeSpec
	genDecl  *ast.GenDecl
}

type NodeVisitor struct {
	declaredInterfaces []declaredInterface
	ctx                context.Context
	lastSeenGenDecl    *ast.GenDecl
}

func NewNodeVisitor(ctx context.Context) *NodeVisitor {
	return &NodeVisitor{
		declaredInterfaces: make([]declaredInterface, 0),
		ctx:                ctx,
	}
}

func (nv *NodeVisitor) DeclaredInterfaces() []declaredInterface {
	return nv.declaredInterfaces
}

func (nv *NodeVisitor) Visit(node ast.Node) ast.Visitor {
	log := zerolog.Ctx(nv.ctx)

	switch n := node.(type) {
	case *ast.BlockStmt:
		// Local type declarations cannot be mocked at package scope.
		return nil
	case *ast.GenDecl:
		// If the next node in the AST is an *ast.TypeSpec, we can rely on this
		// being its parent GenDecl because the walk is done depth-first.
		nv.lastSeenGenDecl = n
	case *ast.TypeSpec:
		log := log.With().
			Str("node-name", n.Name.Name).
			Str("node-type", fmt.Sprintf("%T", n.Type)).
			Logger()

		switch n.Type.(type) {
		case *ast.InterfaceType, *ast.IndexExpr, *ast.IndexListExpr, *ast.SelectorExpr, *ast.Ident:
			log.Debug().Msg("found node with acceptable type for mocking.")
			nv.declaredInterfaces = append(nv.declaredInterfaces, declaredInterface{
				typeSpec: n,
				genDecl:  nv.lastSeenGenDecl,
			})
		default:
			log.Debug().Msg("found node with unacceptable type for mocking. Rejecting.")
		}
	}
	return nv
}
