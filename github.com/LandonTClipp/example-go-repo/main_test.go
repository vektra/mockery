package main

import "testing"

func TestFoo(t *testing.T) {
	m := NewMockFoo(t)
	m.EXPECT().Bar("hello").Return(42)
	if got := m.Bar("hello"); got != 42 {
		t.Errorf("Bar() = %v, want %v", got, 42)
	}
}
