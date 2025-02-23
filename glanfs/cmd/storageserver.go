package cmd

import (
	"github.com/overmighty/glan/glanfs/internal/storageserver"
	"github.com/spf13/cobra"
)

var storageServerCmd = &cobra.Command{
	Use:   "storage-server",
	Short: "Run storage server",
	Long:  "Run storage server",
	Run:   runStorageServer,
}

var storageServerConfig = &storageserver.Config{}

func init() {
	rootCmd.AddCommand(storageServerCmd)
	storageServerCmd.Flags().Uint64Var(&storageServerConfig.CapacityGiB, "capacity-gib", 12, "storage capacity in GiB")
	storageServerCmd.Flags().StringVar(&storageServerConfig.MasterServerAddr, "master", "127.0.0.1:42701", "master server address")
}

func runStorageServer(*cobra.Command, []string) {
	s := &storageserver.StorageServer{Config: storageServerConfig}
	if err := s.Run(); err != nil {
		panic(err)
	}
}
