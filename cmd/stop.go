package cmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-errors/errors"
	ps "github.com/mitchellh/go-ps"
	"github.com/pastelnetwork/gonode/common/cli"
	"github.com/pastelnetwork/gonode/common/log"
	"github.com/pastelnetwork/gonode/common/sys"
	"github.com/pastelnetwork/pastel-utility/configs"
	"github.com/pastelnetwork/pastel-utility/constants"
	"github.com/pastelnetwork/pastel-utility/utils"
)

type stopCommand uint8

const (
	nodeStop stopCommand = iota
	walletStop
	superNodeStop
	allStop
)

func setupStopSubCommand(config *configs.Config,
	stopCommand stopCommand,
	f func(context.Context, *configs.Config) error,
) *cli.Command {

	commonFlags := []*cli.Flag{
		cli.NewFlag("dir", &config.PastelExecDir).SetAliases("d").
			SetUsage(green("Location where to create pastel node directory")).SetValue(config.PastelExecDir),
		cli.NewFlag("work-dir", &config.WorkingDir).SetAliases("w").
			SetUsage(green("Location where to create working directory")).SetValue(config.WorkingDir),
	}

	var commandName, commandMessage string
	var commandFlags []*cli.Flag = commonFlags

	switch stopCommand {
	case nodeStop:
		commandName = "node"
		commandMessage = "Stop node"
	case walletStop:
		commandName = "walletnode"
		commandMessage = "Stop walletnode"
	case superNodeStop:
		commandName = "supernode"
		commandMessage = "Stop supernode"
	case allStop:
		commandName = "all"
		commandMessage = "Stop all"

	default:
		commandName = "all"
		commandMessage = "Stop all"
	}

	subCommand := cli.NewCommand(commandName)
	subCommand.SetUsage(cyan(commandMessage))
	subCommand.AddFlags(commandFlags...)
	if f != nil {
		subCommand.SetActionFunc(func(ctx context.Context, args []string) error {
			ctx, err := configureLogging(ctx, commandMessage, config)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			sys.RegisterInterruptHandler(cancel, func() {
				log.WithContext(ctx).Info("Interrupt signal received. Gracefully shutting down...")
				os.Exit(0)
			})

			log.WithContext(ctx).Info("Stopping...")
			err = f(ctx, config)
			if err != nil {
				return err
			}
			log.WithContext(ctx).Info("Finished successfully!")
			return nil
		})
	}
	return subCommand
}

func setupStopCommand() *cli.Command {
	config := configs.GetConfig()

	stopNodeSubCommand := setupStopSubCommand(config, nodeStop, runStopNodeSubCommand)
	stopWalletSubCommand := setupStopSubCommand(config, walletStop, runStopWalletSubCommand)
	stopSuperNodeSubCommand := setupStopSubCommand(config, superNodeStop, runStopSuperNodeSubCommand)
	stopallSubCommand := setupStopSubCommand(config, allStop, runStopAllSubCommand)

	stopCommand := cli.NewCommand("stop")
	stopCommand.SetUsage(blue("Performs stop of the system for both WalletNode and SuperNodes"))
	stopCommand.AddSubcommands(stopNodeSubCommand)
	stopCommand.AddSubcommands(stopWalletSubCommand)
	stopCommand.AddSubcommands(stopSuperNodeSubCommand)
	stopCommand.AddSubcommands(stopallSubCommand)

	return stopCommand
}

func runStopAllSubCommand(ctx context.Context, config *configs.Config) error {

	var err error
	// *************  Kill process super node  *************
	log.WithContext(ctx).Info("Start stopping supernode process")
	if err = processKill(constants.SuperNode); err != nil {
		return err
	}
	log.WithContext(ctx).Info("The Supernode stopped.")

	// *************  Kill process wallet node  *************
	log.WithContext(ctx).Info("Start stopping walletnode process")
	if err = processKill(constants.WalletNode); err != nil {
		return err
	}
	log.WithContext(ctx).Info("The Walletnode stopped.")

	// *************  Kill process wallet node  *************
	log.WithContext(ctx).Info("Start stopping rqservice process")
	if err = processKill(constants.RQService); err != nil {
		return err
	}
	log.WithContext(ctx).Info("The rqservice stopped.")

	// TODO: Implement Stop node command
	log.WithContext(ctx).Info("Pasteld process kill starting.")
	if _, err = stopPatelCLI(ctx, config); err != nil {
		return err
	}
	log.WithContext(ctx).Info("Pasteld process ended.")

	time.Sleep(10000 * time.Millisecond)

	log.WithContext(ctx).Info("End successfully")

	return nil
}

func runStopNodeSubCommand(ctx context.Context, config *configs.Config) error {

	log.WithContext(ctx).Info("Start stopping Pasteld process")
	if _, err := stopPatelCLI(ctx, config); err != nil {
		return err
	}
	log.WithContext(ctx).Info("Pasteld process ended.")
	log.WithContext(ctx).Info("End successfully")

	return nil
}

func runStopWalletSubCommand(ctx context.Context, config *configs.Config) error {

	var err error

	// *************  Kill process wallet node  *************
	log.WithContext(ctx).Info("Start stopping Walletnode process")

	if err = processKill(constants.WalletNode); err != nil {
		return err
	}
	log.WithContext(ctx).Info("Walletnode process ended.")

	// *************  Kill process rqservice  *************
	log.WithContext(ctx).Info("Start stopping rqservice process")
	if err = processKill(constants.RQService); err != nil {
		return err
	}
	log.WithContext(ctx).Info("The rqservice process ended.")

	// *************  Stop pasteld  *************
	log.WithContext(ctx).Info("Start stopping Pasteld process")

	if _, err = stopPatelCLI(ctx, config); err != nil {
		return err
	}
	log.WithContext(ctx).Info("Pasteld process ended.")

	log.WithContext(ctx).Info("End successfully")

	return nil
}

func runStopSuperNodeSubCommand(ctx context.Context, config *configs.Config) error {
	var err error

	// *************  Kill process super node  *************
	log.WithContext(ctx).Info("Start stopping Supernode process")

	if err = processKill(constants.SuperNode); err != nil {
		return err
	}
	log.WithContext(ctx).Info("Supernode process ended.")

	// *************  Kill process rqservice  *************
	log.WithContext(ctx).Info("Start stopping rqservice process")

	if err = processKill(constants.RQService); err != nil {
		return err
	}
	log.WithContext(ctx).Info("The rqservice process ended.")

	// *************  Stop pastel super node   *************
	log.WithContext(ctx).Info("Start stopping Pasteld process")

	if _, err = stopPatelCLI(ctx, config); err != nil {
		return err
	}
	log.WithContext(ctx).Info("Pasteld process ended.")

	return nil
}

func stopPatelCLI(ctx context.Context, config *configs.Config) (output string, err error) {
	if _, err = runPastelCLI(ctx, config, "stop"); err != nil {
		return "", err
	}

	return "", nil
}

func processKill(toolType constants.ToolType) error {
	execName := constants.ServiceName[toolType][utils.GetOS()]
	proc, err := ps.Processes()
	if err != nil {
		return errors.Errorf("failed to get list process: %v", err)
	}
	pid := 0
	for _, p := range proc {
		if strings.Contains(execName, p.Executable()) {
			pid = p.Pid()
			break
		}
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.Errorf("failed to find %s process: %v", execName, err)
	}

	process.Kill()
	return nil
}
