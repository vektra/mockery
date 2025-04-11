package template_funcs

import (
	"cmp"
	"os"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/exp/constraints"
)

func Exported(s string) string {
	if s == "" {
		return ""
	}
	for _, initialism := range golintInitialisms {
		if strings.ToUpper(s) == initialism {
			return initialism
		}
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}

func ReadFile(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

// Numbers defines the generic constraints of the arithmetic arguments.
type Numbers interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

// Add adds the given numbers.
func Add[T Numbers](i1 T, in ...T) T {
	var sum T = i1
	for _, i := range in {
		sum += i
	}
	return sum
}

// Incr increments the numbers by 1.
func Incr[T Numbers](i T) T {
	return i + 1
}

// Decr decrements the numbers by 1.
func Decr[T Numbers](i T) T {
	return i - 1
}

// Sub subtracts the given numbers.
func Sub[T Numbers](i1 T, in ...T) T {
	var sub T = i1
	for _, i := range in {
		sub -= i
	}
	return sub
}

// Div cumulatively divides the given numbers.
func Div[T Numbers](i1 T, in ...T) T {
	var sub T = i1
	for _, i := range in {
		sub /= i
	}
	return sub
}

// Mod returns the cumulative modulo of the given numbers.
func Mod[T constraints.Integer](i1 T, in ...T) T {
	var sub T = i1
	for _, i := range in {
		sub = sub % i
	}
	return sub
}

// Mul returns the cumulative multiplication of the given numbers.
func Mul[T Numbers](i1 T, in ...T) T {
	var sub T = i1
	for _, i := range in {
		sub *= i
	}
	return sub
}

// Max returns the maximum value.
func Max[T cmp.Ordered](x ...T) T {
	return slices.Max(x)
}

// Min returns the minimum value.
func Min[T cmp.Ordered](x ...T) T {
	return slices.Min(x)
}

// FirstIsLower returns whether or not the string's first character is lowercase.
// If the string is empty, false is returned. If the first character is a non-alphabetic
// character, false is returned.
func FirstIsLower(s string) bool {
	first := rune(s[0])
	if len(s) == 0 || !unicode.IsLetter(first) {
		return false
	}
	return !unicode.IsUpper(first)
}
