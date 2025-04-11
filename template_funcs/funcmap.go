// Package shared provides variables/objects that need to be shared
// across multiple packages. The main purpose is to resolve cyclical imports
// arising from multiple packages needing to share common utilies.
package template_funcs

import (
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
)

// This list comes from the golint codebase. Golint will complain about any of
// these being mixed-case, like "Id" instead of "ID".
var golintInitialisms = []string{
	"ACL", "API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS",
	"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "TCP", "TLS", "TTL", "UDP", "UI", "UID", "UUID", "URI",
	"URL", "UTF8", "VM", "XML", "XMPP", "XSRF", "XSS",
}

//nolint:predeclared
var FuncMap = template.FuncMap{
	// String inspection and manipulation. Note that the first argument is replaced
	// as the last argument in some functions in order to support chained
	// template pipelines.
	"contains":     func(substr string, s string) bool { return strings.Contains(s, substr) },
	"hasPrefix":    func(prefix string, s string) bool { return strings.HasPrefix(s, prefix) },
	"hasSuffix":    func(suffix string, s string) bool { return strings.HasSuffix(s, suffix) },
	"join":         func(sep string, elems []string) string { return strings.Join(elems, sep) },
	"replace":      func(old string, new string, n int, s string) string { return strings.Replace(s, old, new, n) },
	"replaceAll":   func(old string, new string, s string) string { return strings.ReplaceAll(s, old, new) },
	"split":        func(sep string, s string) []string { return strings.Split(s, sep) },
	"splitAfter":   func(sep string, s string) []string { return strings.SplitAfter(s, sep) },
	"splitAfterN":  func(sep string, n int, s string) []string { return strings.SplitAfterN(s, sep, n) },
	"trim":         func(cutset string, s string) string { return strings.Trim(s, cutset) },
	"trimLeft":     func(cutset string, s string) string { return strings.TrimLeft(s, cutset) },
	"trimPrefix":   func(prefix string, s string) string { return strings.TrimPrefix(s, prefix) },
	"trimRight":    func(cutset string, s string) string { return strings.TrimRight(s, cutset) },
	"trimSpace":    strings.TrimSpace,
	"trimSuffix":   func(suffix string, s string) string { return strings.TrimSuffix(s, suffix) },
	"lower":        strings.ToLower,
	"upper":        strings.ToUpper,
	"camelcase":    xstrings.ToCamelCase,
	"snakecase":    xstrings.ToSnakeCase,
	"kebabcase":    xstrings.ToKebabCase,
	"firstIsLower": FirstIsLower,
	"firstLower":   xstrings.FirstRuneToLower,
	"firstUpper":   xstrings.FirstRuneToUpper,
	"exported":     Exported,

	// Regular expression matching
	"matchString": regexp.MatchString,
	"quoteMeta":   regexp.QuoteMeta,

	// Filepath manipulation
	"base":     filepath.Base,
	"clean":    filepath.Clean,
	"dir":      filepath.Dir,
	"readFile": ReadFile,

	// Basic access to reading environment variables
	"expandEnv": os.ExpandEnv,
	"getenv":    os.Getenv,

	/*******
	* MATH *
	********/
	// int
	"add":  Add[int],
	"decr": Decr[int],
	"div":  Div[int],
	"incr": Incr[int],
	"min":  Min[int],
	"mod":  Mod[int],
	"mul":  Mul[int],
	"sub":  Sub[int],

	// float64
	"ceil":  math.Ceil,
	"floor": math.Floor,
	"round": math.Round,

	// rand
	"randInt": rand.Int,
}
