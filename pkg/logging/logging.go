package logging

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/stackerr"
	"golang.org/x/term"
)

const (
	LogKeyBaseDir       = "base-dir"
	LogKeyDir           = "dir"
	LogKeyDryRun        = "dry-run"
	LogKeyFile          = "file"
	LogKeyInterface     = "interface"
	LogKeyImport        = "import"
	LogKeyPath          = "path"
	LogKeyQualifiedName = "qualified-name"
	LogKeyPackageName   = "package-name"
	_defaultSemVer      = "v0.0.0-dev"
)

// SemVer is the version of mockery at build time.
var SemVer = ""
var ErrPkgNotExist = errors.New("package does not exist")

func GetSemverInfo() string {
	if SemVer != "" {
		return SemVer
	}
	version, ok := debug.ReadBuildInfo()
	if ok && version.Main.Version != "(devel)" && version.Main.Version != "" {
		return version.Main.Version
	}
	return _defaultSemVer
}

func getMinorSemver(semver string) string {
	split := strings.Split(semver, ".")
	return strings.Join(split[0:2], ".")
}

// GetMinorSemver returns the semantic version up to and including the minor version.
func GetMinorSemver() string {
	return getMinorSemver(GetSemverInfo())
}

func DocsURL(relativePath string) string {
	if string(relativePath[0]) != "/" {
		relativePath = "/" + relativePath
	}
	return fmt.Sprintf("https://vektra.github.io/mockery/%s%s", GetMinorSemver(), relativePath)
}

type timeHook struct{}

func (t timeHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Time("time", time.Now())
}

func GetLogger(levelStr string) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return zerolog.Logger{}, stackerr.NewStackErrf(err, "Couldn't parse log level")
	}
	out := os.Stderr
	writer := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: time.RFC822,
	}
	if !term.IsTerminal(int(out.Fd())) || os.Getenv("TERM") == "dumb" {
		writer.NoColor = true
	}
	log := zerolog.New(writer).
		Hook(timeHook{}).
		Level(level).
		With().
		Str("version", GetSemverInfo()).
		Logger()

	return log, nil
}
