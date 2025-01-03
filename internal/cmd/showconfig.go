package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v3/internal"
	"github.com/vektra/mockery/v3/internal/logging"
	"github.com/vektra/mockery/v3/internal/stackerr"
	"gopkg.in/yaml.v3"
)

func NewShowConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "showconfig",
		Short: "Show the yaml config",
		Long:  `Print out a yaml representation of the yaml config file. This does not show config from exterior sources like CLI, environment etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := getConfig(nil, nil)
			if err != nil {
				return err
			}
			return showConfig(cmd, args, cfg, os.Stdout)
		},
	}
}

func showConfig(
	cmd *cobra.Command,
	args []string,
	v *viper.Viper,
	outputter io.Writer,
) error {
	ctx := context.Background()
	config, err := pkg.NewConfigFromViper(v)
	if err != nil {
		return stackerr.NewStackErrf(err, "failed to unmarshal config")
	}
	log, err := logging.GetLogger(config.LogLevel)
	if err != nil {
		return fmt.Errorf("getting logger: %w", err)
	}
	ctx = log.WithContext(ctx)
	if err := config.Initialize(ctx); err != nil {
		return err
	}
	cfgMap, err := config.CfgAsMap(ctx)
	if err != nil {
		panic(err)
	}

	encoder := yaml.NewEncoder(outputter)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfgMap); err != nil {
		return stackerr.NewStackErrf(err, "failed to marshal yaml")
	}

	log.Info().Msgf("Using config: %s", config.Config)

	return nil
}