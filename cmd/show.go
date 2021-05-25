package cmd

import (
	"context"
	"os"

	"github.com/pastelnetwork/pastel-utility/configs"
	"github.com/pastelnetwork/gonode/common/cli"
	"github.com/pastelnetwork/gonode/common/log"
	"github.com/pastelnetwork/gonode/common/sys"
)

func setupShowCommand() *cli.Command {
	config := configs.New()

	// define flags here
	var showFlag string

	showCommand := cli.NewCommand("show")
	showCommand.SetUsage("usage")
	showCommandFlags := []*cli.Flag{
		cli.NewFlag("flag-name", &showFlag),
	}
	showCommand.AddFlags(showCommandFlags...)
	addLogFlags(showCommand, config)

	showCommand.SetActionFunc(func(ctx context.Context, args []string) error {
		ctx, err := configureLogging("showcommand", config, ctx)
		if err != nil {
			return err
		}

		log.Info("flag-name: ", showFlag)

		return runShow(ctx, config)
	})
	return showCommand
}

func runShow(ctx context.Context, config *configs.Config) error {
	log.WithContext(ctx).Info("Show")
	defer log.WithContext(ctx).Info("End")

	configJson, err := config.String()
	if err != nil {
		return err
	}
	log.WithContext(ctx).Infof("Config: %s", configJson)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sys.RegisterInterruptHandler(cancel, func() {
		log.WithContext(ctx).Info("Interrupt signal received. Gracefully shutting down...")
		os.Exit(0)
	})

	// actions to run goes here

	return nil

}
