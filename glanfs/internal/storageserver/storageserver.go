package storageserver

import (
	"fmt"
	"github.com/overmighty/glan/glanfs/api/storageapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"log/slog"
	"net"
)

type Config struct {
	CapacityGiB      uint64
	MasterServerAddr string
}

type StorageServer struct {
	Config *Config

	storage *storage

	conn *common.Conn[*storageapi.Response, *storageapi.Request]
}

func (s *StorageServer) capacityBlocks() uint64 {
	return s.Config.CapacityGiB * 1024 * 1024 * 1024 / common.BlockSize
}

func (s *StorageServer) Run() error {
	slog.Debug("Running storage server", "capacity_blocks", s.capacityBlocks())

	s.storage = newStorage()

	conn, err := net.Dial("tcp", s.Config.MasterServerAddr)
	if err != nil {
		return fmt.Errorf("storageserver: failed to connect to master server: %w", err)
	}
	s.conn = common.NewConn[*storageapi.Response, *storageapi.Request](conn)

	s.serve()
	return nil
}
