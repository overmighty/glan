package client

import (
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"log/slog"
	"net"
	"os"
	"sync"
)

type Config struct {
	MountPoint       string
	DebugFuse        bool
	MasterServerAddr string
}

type Client struct {
	Config *Config

	connMu sync.Mutex
	conn   *common.Conn[*fsapi.Request, *fsapi.Response]
}

func (c *Client) Run() error {
	slog.Debug("Running client")

	if _, err := os.Stat(c.Config.MountPoint); os.IsNotExist(err) {
		if err := os.MkdirAll(c.Config.MountPoint, 0700); err != nil {
			return err
		}
	}

	opts := &fs.Options{}
	opts.Name = "glanfs"
	opts.Debug = c.Config.DebugFuse

	glanfsRoot := &glanfsNode{client: c, id: 1}

	server, err := fs.Mount(c.Config.MountPoint, glanfsRoot, opts)
	if err != nil {
		return fmt.Errorf("glanfs/client: failed to mount: %s", err.Error())
	}
	defer func(server *fuse.Server) {
		if err := server.Unmount(); err != nil {
			slog.Error("Failed to unmount", "err", err)
		}
	}(server)

	conn, err := net.Dial("tcp", c.Config.MasterServerAddr)
	if err != nil {
		return fmt.Errorf("glanfs/client: failed to connect to master server: %s", err.Error())
	}
	c.conn = common.NewConn[*fsapi.Request, *fsapi.Response](conn)

	server.Wait()
	return nil
}
