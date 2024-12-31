package iface_new_type_test

import (
	"testing"
)

func TestUsage(t *testing.T) {
	interface1 := NewMockInterface1(t)
	interface1.EXPECT().Method1().Return()
	interface1.Method1()

	interface2 := NewMockInterface2(t)
	interface2.EXPECT().Method1().Return()
	interface2.Method1()

	interface3 := NewMockInterface3(t)
	interface3.EXPECT().Method1().Return()
	interface3.Method1()
}
