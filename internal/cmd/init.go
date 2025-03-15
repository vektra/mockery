package cmd

import (
	"context"
	"os"

	"github.com/chigopher/pathlib"
	"github.com/spf13/cobra"
	"github.com/vektra/mockery/v3/config"
	"github.com/vektra/mockery/v3/internal/logging"
	"gopkg.in/yaml.v3"
)

func addr[T any](v T) *T {
	return &v
}

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [module_name]",
		Short: "Generate a basic .mockery.yml file",
		Long:  `This command generates a basic .mockery.yml file that can be used as a starting point for your config.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := ".mockery.yml"
			log, err := logging.GetLogger("debug")
			if err != nil {
				os.Exit(1)
			}
			moduleName := args[0]
			log.Info().Str("file", filename).Msg("writing to file")
			defer log.Info().Msg("done")

			ctx := log.WithContext(context.Background())
			k, err := config.NewDefaultKoanf(ctx)
			if err != nil {
				log.Err(err).Msg("failed getting koanf")
				os.Exit(1)
			}
			rootConf := &config.RootConfig{}
			if err := k.Unmarshal("", rootConf); err != nil {
				log.Err(err).Msg("failed to unmarshal koanf")
				os.Exit(1)
			}
			rootConf.Packages = map[string]*config.PackageConfig{
				moduleName: {
					Config: &config.Config{
						All: addr(true),
					},
					Interfaces: map[string]*config.InterfaceConfig{},
				},
			}

			outFile := pathlib.NewPath(filename)
			f, err := outFile.OpenFile(os.O_RDWR | os.O_CREATE | os.O_EXCL)
			if err != nil {
				log.Err(err).Msg("failed to open file")
				os.Exit(1)
			}
			defer f.Close()

			encoder := yaml.NewEncoder(f)
			defer encoder.Close()
			encoder.SetIndent(2)
			if err := encoder.Encode(rootConf); err != nil {
				log.Err(err).Msg("failed to encode")
				os.Exit(1)
			}
		},
	}
}
