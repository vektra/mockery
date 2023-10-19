package registry

import (
	"testing"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New("../../pkg/moq/testpackages/example", "")
	}
}
