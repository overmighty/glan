package masterserver

import (
	"fmt"
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"github.com/overmighty/glan/glanfs/api/storageapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"log/slog"
	"net"
)

type Config struct {
	ClientListenerAddr        string
	StorageServerListenerAddr string
}

type MasterServer struct {
	Config *Config

	fsTable     *fsTable
	storageList *storageList

	clientLn  net.Listener
	storageLn net.Listener
}

func (m *MasterServer) Run() error {
	slog.Debug("Running master server")

	m.fsTable = newFSTable()
	rootNode := newDirNode("/")
	if err := m.fsTable.addNode(0, rootNode); err != 0 {
		err.String()
		return fmt.Errorf("failed to create root node: %s", err)
	}

	m.storageList = newStorageList()

	var err error
	if m.clientLn, err = net.Listen("tcp", m.Config.ClientListenerAddr); err != nil {
		return err
	}
	if m.storageLn, err = net.Listen("tcp", m.Config.StorageServerListenerAddr); err != nil {
		return err
	}

	go m.listenForStorageServers()
	m.listenForClients()
	return nil
}

func (m *MasterServer) listenForClients() {
	for {
		conn, err := m.clientLn.Accept()
		if err != nil {
			slog.Error("Failed to accept client conn", "err", err)
			continue
		}
		slog.Debug("Accepted client conn", "remote_addr", conn.RemoteAddr())

		c := &clientConn{
			server: m,
			conn:   common.NewConn[*fsapi.Response, *fsapi.Request](conn),
		}
		go c.serve()
	}
}

func (m *MasterServer) listenForStorageServers() {
	for {
		conn, err := m.storageLn.Accept()
		if err != nil {
			slog.Error("Failed to accept storage server conn", "err", err)
			continue
		}
		slog.Debug("Accepted storage server conn", "remote_addr", conn.RemoteAddr())

		s := &storageServerConn{
			server: m,
			conn:   common.NewConn[*storageapi.Request, *storageapi.Response](conn),
		}
		resp := s.getCapacity()
		s.numFreeBlocks = resp.GetNumBlocks()
		m.storageList.addStorage(s)
	}
}
