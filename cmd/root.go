package cmd

import (
	"embed"
	"github.com/spf13/cobra"
	"nexus3-fsnotify/cmd/version"
)

func GetRootCmd(staticAssets embed.FS, args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "cmdb",
		Short:             "cmdb  interface.",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long:              `The new generation of CMDB`,
	}
	rootCmd.AddCommand(ServerCommand(staticAssets, args))
	rootCmd.AddCommand(version.VersionCommand(args))

	return rootCmd
}
