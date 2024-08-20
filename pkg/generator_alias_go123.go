//go:build go1.23

package pkg

import (
	"context"
	"fmt"
	"go/types"
	"strings"
)

func (g *Generator) renderTypeAlias(ctx context.Context, t *types.Alias) string {
	name := g.getPackageScopedType(ctx, t.Obj())
	if t.TypeArgs() == nil || t.TypeArgs().Len() == 0 {
		return name
	}
	args := make([]string, 0, t.TypeArgs().Len())
	for i := 0; i < t.TypeArgs().Len(); i++ {
		arg := t.TypeArgs().At(i)
		args = append(args, g.renderType(ctx, arg))
	}
	return fmt.Sprintf("%s[%s]", name, strings.Join(args, ","))
}
