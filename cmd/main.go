package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/at-ishikawa/local-package-updater/internal/packagemanager"
	"github.com/spf13/cobra"
)

func main() {
	exitCode, err := runMain()
	if err != nil {
		slog.Error("error", "err", err)
	}
	os.Exit(exitCode)
}

func runMain() (int, error) {
	ctx := context.Background()

	var isDebug bool
	rootCommand := &cobra.Command{
		Use:              "local-package-updater",
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel := slog.LevelInfo
			if isDebug {
				logLevel = slog.LevelDebug
			}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: logLevel,
			}))
			slog.SetDefault(logger)
			return nil
		},
	}
	rootCommand.PersistentFlags().BoolVarP(&isDebug, "debug", "d", false, "enable a debug mode")

	rootCommand.AddCommand(&cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginManagers := []packagemanager.Plugin{
				packagemanager.NewGeneralManager(packagemanager.CLIArgs{
					"gcloud", "components", "update",
				}, false),
				packagemanager.NewAptManager(true),
				packagemanager.NewGeneralManager(packagemanager.CLIArgs{
					"fish", "-c", "fisher update",
				}, false),
				packagemanager.NewGeneralManager(packagemanager.CLIArgs{
					"kubectl", "krew", "upgrade",
				}, false),
				packagemanager.NewGeneralManager(packagemanager.CLIArgs{
					"brew", "update",
				}, false),
			}

			for _, isSudoRequired := range []bool{true, false} {
				for _, pluginManager := range pluginManagers {
					if pluginManager.IsSudoRequired() != isSudoRequired {
						continue
					}
					if ok := pluginManager.IsCommandInstalled(); !ok {
						continue
					}

					if err := pluginManager.Update(); err != nil {
						return fmt.Errorf("failed to update %w", err)
					}
				}
			}

			return nil
		},
	})

	if err := rootCommand.ExecuteContext(ctx); err != nil {
		return 1, err
	}
	return 0, nil
}
