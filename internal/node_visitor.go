package pkg

import (
	"context"
	"fmt"
	"go/ast"

	"github.com/rs/zerolog"
)

type NodeVisitor struct {
	declaredInterfaces []string
	ctx                context.Context
}

func NewNodeVisitor(ctx context.Context) *NodeVisitor {
	return &NodeVisitor{
		declaredInterfaces: make([]string, 0),
		ctx:                ctx,
	}
}

func (nv *NodeVisitor) DeclaredInterfaces() []string {
	return nv.declaredInterfaces
}

func (nv *NodeVisitor) add(ctx context.Context, n *ast.TypeSpec) {
	log := zerolog.Ctx(ctx)
	log.Debug().
		Str("node-name", n.Name.Name).
		Str("node-type", fmt.Sprintf("%T", n.Type)).
		Msg("found type declaration that is a possible interface")
	nv.declaredInterfaces = append(nv.declaredInterfaces, n.Name.Name)
}

func (nv *NodeVisitor) Visit(node ast.Node) ast.Visitor {
	log := zerolog.Ctx(nv.ctx)

	switch n := node.(type) {
	case *ast.TypeSpec:
		log := log.With().
			Str("node-name", n.Name.Name).
			Str("node-type", fmt.Sprintf("%T", n.Type)).
			Logger()

		switch n.Type.(type) {
		case *ast.InterfaceType, *ast.IndexExpr, *ast.IndexListExpr:
			nv.add(nv.ctx, n)
		default:
			log.Debug().Msg("found node with unacceptable type for mocking. Rejecting.")
		}
	}
	return nv
}
