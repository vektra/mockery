package cmd

import (
	"fmt"
	"os"

	"github.com/go-errors/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger zerolog.Logger

type timestampHook struct{}

func (t timestampHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	e.Timestamp()
}

func maybeExit(err error) {
	printStack(err)
	if err != nil {
		os.Exit(1)
	}
}

func newViper() *viper.Viper {
	v := viper.New()
	v.SetConfigType("env")
	v.SetConfigName("mockery-tools")
	v.AddConfigPath(".")
	v.AddConfigPath("../")
	v.SetEnvPrefix("MOCKERYTOOLS")
	maybeExit(v.ReadInConfig())
	return v
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "mockery_tools [command]",
	}

	logger = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr,
	}).Hook(timestampHook{})

	subCommands := []func(v *viper.Viper) (*cobra.Command, error){
		NewTagCmd,
	}
	for _, CommandFunc := range subCommands {
		subCmd, err := CommandFunc(newViper())
		if err != nil {
			panic(err)
		}
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func printStack(err error) {
	if err == nil {
		return
	}
	newErr, ok := err.(*errors.Error)
	if ok {
		fmt.Fprintf(os.Stderr, "%v\n", newErr.ErrorStack())
	} else {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
