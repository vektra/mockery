package iface_new_type

import (
	"testing"
)

func TestUsage(t *testing.T) {
	interface1 := NewMockInterface1(t)
	interface1.EXPECT().Method1().Return()
	interface1.Method1()
}
