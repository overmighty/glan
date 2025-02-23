package cmd

import (
	"github.com/overmighty/glan/glanfs/internal/client"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run client",
	Long:  "Run client",
	Run:   runClient,
}

var clientConfig = &client.Config{}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVar(&clientConfig.MountPoint, "mountpoint", "/mnt", "mount point")
	clientCmd.Flags().BoolVar(&clientConfig.DebugFuse, "debug-fuse", false, "print FUSE debug info")
	clientCmd.Flags().StringVar(&clientConfig.MasterServerAddr, "master", "127.0.0.1:42700", "master server address")
}

func runClient(*cobra.Command, []string) {
	c := &client.Client{Config: clientConfig}
	if err := c.Run(); err != nil {
		panic(err)
	}
}
