package context

import (
	"context"
)

type CollideWithStdLib interface {
	NewClient(ctx context.Context)
}
