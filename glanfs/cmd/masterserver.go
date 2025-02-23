package cmd

import (
	"github.com/overmighty/glan/glanfs/internal/masterserver"
	"github.com/spf13/cobra"
)

var masterServerCmd = &cobra.Command{
	Use:   "master-server",
	Short: "Run master server",
	Long:  "Run master server",
	Run:   runMasterServer,
}

var masterServerConfig = &masterserver.Config{}

func init() {
	rootCmd.AddCommand(masterServerCmd)
	masterServerCmd.Flags().StringVar(&masterServerConfig.ClientListenerAddr, "listen-client", ":42700", "client listener address")
	masterServerCmd.Flags().StringVar(&masterServerConfig.StorageServerListenerAddr, "listen-storage", ":42701", "storage server listener address")
}

func runMasterServer(*cobra.Command, []string) {
	c := &masterserver.MasterServer{Config: masterServerConfig}
	if err := c.Run(); err != nil {
		panic(err)
	}
}
