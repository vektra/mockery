package registry_test

import (
	"testing"

	"github.com/matryer/moq/internal/registry"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		registry.New("../../pkg/moq/testpackages/example", "")
	}
}
