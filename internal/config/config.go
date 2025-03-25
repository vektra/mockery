package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/chigopher/pathlib"
)

func FindConfig() (*pathlib.Path, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current working directory: %w", err)
	}
	currentPath := pathlib.NewPath(cwd)
	for len(currentPath.Parts()) != 1 {
		for _, confName := range []string{".mockery.yaml", ".mockery.yml"} {
			configPath := currentPath.Join(confName)
			isFile, err := configPath.Exists()
			if err != nil {
				return nil, fmt.Errorf("checking if %s is file: %w", configPath.String(), err)
			}
			if isFile {
				return configPath, nil
			}
		}
		currentPath = currentPath.Parent()
	}
	return nil, errors.New("mockery config file not found")
}
