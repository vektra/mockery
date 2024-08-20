//go:build !go1.23

package pkg

import (
	"context"
	"go/types"
)

func (g *Generator) renderTypeAlias(ctx context.Context, t *types.Alias) string {
	return g.getPackageScopedType(ctx, t.Obj())
}
