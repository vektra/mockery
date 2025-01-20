package cmd

import (
	"context"
	"fmt"

	koanfYAML "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/vektra/mockery/v3/config"
	"github.com/vektra/mockery/v3/internal/logging"
)

func NewShowConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "showconfig",
		Short: "Show the yaml config",
		Long:  `Print out a yaml representation of the yaml config file. This does not show config from exterior sources like CLI, environment etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log, err := logging.GetLogger("debug")
			if err != nil {
				return err
			}

			ctx := log.WithContext(context.Background())
			conf, _, err := config.NewRootConfig(ctx, cmd.Parent().PersistentFlags())
			if err != nil {
				return err
			}

			k := koanf.New("|")
			if err := k.Load(structs.Provider(conf, "koanf"), nil); err != nil {
				log.Err(err).Msg("failed to load config")
				return err
			}
			b, _ := k.Marshal(koanfYAML.Parser())
			fmt.Println(string(b))

			return nil
		},
	}
}
