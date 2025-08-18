package cmd

import (
	"embed"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"nexus3-fsnotify/src/server"
	"nexus3-fsnotify/src/utils/homedir"
	"os"
	"path/filepath"
)

func ServerCommand(staticAssets embed.FS, args []string) *cobra.Command {
	config := &server.CmdbConfig{}
	serverCommand := &cobra.Command{
		Use:               "server",
		Short:             "run cmdb server ",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Long:              `The new generation of CMDB`,
		Run: func(cmd *cobra.Command, args []string) {
			klog.Info("cmdb start ...")
			config.StaticAssets = staticAssets
			if config.DataRoot == "" {
				config.DataRoot = filepath.Join(homedir.HomeDir(), ".cmdb")
			}
			s, err := server.NewCmdb(config)
			if err != nil {
				klog.Errorf("new cmdb config error:%s", err)
				os.Exit(-1)
			}
			s.ListenAndServe()
		},
	}
	AddServerCommandFlag(serverCommand, config)
	return serverCommand
}

func AddServerCommandFlag(serverCommand *cobra.Command, config *server.CmdbConfig) {
	serverCommand.Flags().Int32VarP(&config.Port, "port", "p", 9090, "http server port")
	serverCommand.Flags().Int32Var(&config.ProxyPort, "proxy-port", 9891, "http server proxy port")
	serverCommand.Flags().StringVar(&config.NexusMetadata.URL, "nexus-url", "http://localhost:8081", "nexus api url")
	serverCommand.Flags().StringVar(&config.NexusMetadata.Username, "nexus-username", "admin", "nexus api username")
	serverCommand.Flags().StringVar(&config.NexusMetadata.Password, "nexus-password", "admin", "nexus api password")
}
