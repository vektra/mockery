package cmd

import (
	"context"
	"fmt"

	koanfYAML "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	pkg "github.com/vektra/mockery/v3/internal"
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
			conf, _, err := pkg.NewRootConfig(ctx, nil, cmd.Parent().PersistentFlags())
			if err != nil {
				return err
			}

			k := koanf.New("|")
			k.Load(structs.Provider(conf, "koanf"), nil)
			b, _ := k.Marshal(koanfYAML.Parser())
			fmt.Println(string(b))

			return nil
		},
	}
}
