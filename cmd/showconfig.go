package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/v2/pkg/config"
	"github.com/vektra/mockery/v2/pkg/logging"
	"github.com/vektra/mockery/v2/pkg/stackerr"
	"gopkg.in/yaml.v2"
)

func NewShowConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "showconfig",
		Short: "Show the yaml config",
		Long:  `Print out a yaml representation of the yaml config file. This does not show config from exterior sources like CLI, environment etc.`,
		RunE:  func(cmd *cobra.Command, args []string) error { return showConfig(cmd, args, viperCfg, os.Stdout) },
	}
}

func showConfig(
	cmd *cobra.Command,
	args []string,
	v *viper.Viper,
	outputter io.Writer,
) error {
	if v == nil {
		v = viperCfg
	}
	ctx := context.Background()
	config, err := config.NewConfigFromViper(v)
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
	out, err := yaml.Marshal(cfgMap)
	if err != nil {
		return stackerr.NewStackErrf(err, "failed to marshal yaml")
	}

	log.Info().Msgf("Using config: %s", config.Config)

	fmt.Fprintf(outputter, "%s", string(out))
	return nil
}
