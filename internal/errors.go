package internal

import "fmt"

var (
	ErrNoConfigFile         = fmt.Errorf("no config file exists")
	ErrNoGoFilesFoundInRoot = fmt.Errorf("no go files found in root search path")
	ErrPkgNotFound          = fmt.Errorf("package not found in config")
	ErrGoModNotFound        = fmt.Errorf("no go.mod file found")
	ErrGoModInvalid         = fmt.Errorf("go.mod file has no module line")
)
