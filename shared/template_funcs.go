// Package shared provides variables/objects that need to be shared
// across multiple packages. The main purpose is to resolve cyclical imports
// arising from multiple packages needing to share common utilies.
package shared

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
)

//nolint:predeclared
var StringManipulationFuncs = template.FuncMap{
	// String inspection and manipulation. Note that the first argument is replaced
	// as the last argument in some functions in order to support chained
	// template pipelines.
	"contains":    func(substr string, s string) bool { return strings.Contains(s, substr) },
	"hasPrefix":   func(prefix string, s string) bool { return strings.HasPrefix(s, prefix) },
	"hasSuffix":   func(suffix string, s string) bool { return strings.HasSuffix(s, suffix) },
	"join":        func(sep string, elems []string) string { return strings.Join(elems, sep) },
	"replace":     func(old string, new string, n int, s string) string { return strings.Replace(s, old, new, n) },
	"replaceAll":  func(old string, new string, s string) string { return strings.ReplaceAll(s, old, new) },
	"split":       func(sep string, s string) []string { return strings.Split(s, sep) },
	"splitAfter":  func(sep string, s string) []string { return strings.SplitAfter(s, sep) },
	"splitAfterN": func(sep string, n int, s string) []string { return strings.SplitAfterN(s, sep, n) },
	"trim":        func(cutset string, s string) string { return strings.Trim(s, cutset) },
	"trimLeft":    func(cutset string, s string) string { return strings.TrimLeft(s, cutset) },
	"trimPrefix":  func(prefix string, s string) string { return strings.TrimPrefix(s, prefix) },
	"trimRight":   func(cutset string, s string) string { return strings.TrimRight(s, cutset) },
	"trimSpace":   strings.TrimSpace,
	"trimSuffix":  func(suffix string, s string) string { return strings.TrimSuffix(s, suffix) },
	"lower":       strings.ToLower,
	"upper":       strings.ToUpper,
	"camelcase":   xstrings.ToCamelCase,
	"snakecase":   xstrings.ToSnakeCase,
	"kebabcase":   xstrings.ToKebabCase,
	"firstLower":  xstrings.FirstRuneToLower,
	"firstUpper":  xstrings.FirstRuneToUpper,

	// Regular expression matching
	"matchString": regexp.MatchString,
	"quoteMeta":   regexp.QuoteMeta,

	// Filepath manipulation
	"base":  filepath.Base,
	"clean": filepath.Clean,
	"dir":   filepath.Dir,

	// Basic access to reading environment variables
	"expandEnv": os.ExpandEnv,
	"getenv":    os.Getenv,

	// Arithmetic
	"add": func(i1, i2 int) int { return i1 + i2 },
}
