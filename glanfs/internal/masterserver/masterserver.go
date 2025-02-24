package masterserver

import (
	"context"
	"fmt"
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"github.com/overmighty/glan/glanfs/api/storageapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"log/slog"
	"net"
)

var meter = otel.Meter("github.com/overmighty/glan/glanfs/internal/masterserver")

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

	clientCounter       metric.Int64UpDownCounter
	responseTimeHist    metric.Int64Histogram
	bytesReadCounter    metric.Int64UpDownCounter
	bytesWrittenCounter metric.Int64UpDownCounter
}

func (m *MasterServer) Run() error {
	slog.Debug("Running master server")

	err := m.initInstrumentation()
	if err != nil {
		return fmt.Errorf("masterserver: failed to init instrumentation: %w", err)
	}

	m.fsTable = newFSTable()
	rootNode := newDirNode("/")
	if err := m.fsTable.addNode(0, rootNode); err != 0 {
		return fmt.Errorf("masterserver: failed to create root node: %v", err)
	}

	m.storageList, err = newStorageList()
	if err != nil {
		return fmt.Errorf("masterserver: failed to init storage list: %w", err)
	}

	if m.clientLn, err = net.Listen("tcp", m.Config.ClientListenerAddr); err != nil {
		return fmt.Errorf("masterserver: failed to listen on %s: %w", m.Config.ClientListenerAddr, err)
	}
	if m.storageLn, err = net.Listen("tcp", m.Config.StorageServerListenerAddr); err != nil {
		return fmt.Errorf("masterserver: failed to listen on %s: %w", m.Config.StorageServerListenerAddr, err)
	}

	go m.listenForStorageServers()
	m.listenForClients()
	return nil
}

func (m *MasterServer) initInstrumentation() error {
	var err error
	m.clientCounter, err = meter.Int64UpDownCounter(
		"glanfs.masterserver.clients",
		metric.WithUnit("{client}"),
	)
	if err != nil {
		return err
	}

	m.responseTimeHist, err = meter.Int64Histogram(
		"glanfs.masterserver.response_time",
		metric.WithUnit("ns"),
	)
	if err != nil {
		return err
	}

	m.bytesReadCounter, err = meter.Int64UpDownCounter(
		"glanfs.masterserver.bytes_read",
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	m.bytesWrittenCounter, err = meter.Int64UpDownCounter(
		"glanfs.masterserver.bytes_written",
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

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
		m.clientCounter.Add(context.Background(), 1)

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
