package masterserver

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"log/slog"
)

type clientConn struct {
	server *MasterServer

	conn *common.Conn[*fsapi.Response, *fsapi.Request]
}

func (c *clientConn) serve() {
	defer c.conn.Close()

	for {
		var req fsapi.Request
		//startTime := time.Now()
		err := c.conn.ReadMessage(&req)
		//endTime := time.Now()
		//slog.Debug("clientConn", "wait_time_ns", endTime.Sub(startTime).Nanoseconds())
		if err != nil {
			slog.Error("Failed to read message", "err", err, "remote_addr", c.conn.RemoteAddr())
			return
		}

		switch req.WhichBody() {
		case fsapi.Request_Lookup_case:
			c.handleLookup(&req)
		case fsapi.Request_Create_case:
			c.handleCreate(&req)
		case fsapi.Request_Readdir_case:
			c.handleReaddir(&req)
		case fsapi.Request_Getattr_case:
			c.handleGetattr(&req)

		case fsapi.Request_Write_case:
			c.handleWrite(&req)
		case fsapi.Request_Read_case:
			c.handleRead(&req)

		default:
			slog.Error("Received invalid request from client", "remote_addr", c.conn.RemoteAddr())
			return
		}
	}
}

func (c *clientConn) handleLookup(req *fsapi.Request) {
	lookup := req.GetLookup()

	parent, ok := c.server.fsTable.getDir(lookup.GetParentId())
	if !ok {
		c.respondError(fsapi.Error_ERROR_ENOENT)
		return
	}

	node, ok := parent.getEntry(lookup.GetName())
	if !ok {
		c.respondError(fsapi.Error_ERROR_ENOENT)
		return
	}

	c.respondLookup(node.getID(), node.getType())
}

func (c *clientConn) handleCreate(req *fsapi.Request) {
	create := req.GetCreate()

	node := newFileNode(create.GetName())
	if err := c.server.fsTable.addNode(create.GetParentId(), node); err != 0 {
		c.respondError(err)
	}

	c.respondCreate(node.getID())
}

func (c *clientConn) handleReaddir(req *fsapi.Request) {
	readdir := req.GetReaddir()

	node, ok := c.server.fsTable.getDir(readdir.GetId())
	if !ok {
		c.respondError(fsapi.Error_ERROR_ENOENT)
		return
	}

	var nodes []*fsapi.Node
	for _, child := range node.entriesByName {
		var childPB fsapi.Node
		childPB.SetId(child.getID())
		childPB.SetType(child.getType())
		childPB.SetName(child.getName())
		nodes = append(nodes, &childPB)
	}
	c.respondReaddir(nodes)
}

func (c *clientConn) handleGetattr(req *fsapi.Request) {
	getattr := req.GetGetattr()

	node, ok := c.server.fsTable.getNode(getattr.GetId())
	if !ok {
		c.respondError(fsapi.Error_ERROR_ENOENT)
		return
	}

	file, ok := node.(*fileNode)
	if !ok {
		c.respondGetattr(0)
		return
	}
	c.respondGetattr(file.size())
}

func (c *clientConn) handleWrite(req *fsapi.Request) {
	write := req.GetWrite()

	size := uint64(len(write.GetData()))
	offset := write.GetOffset()

	if offset%common.BlockSize+size > common.BlockSize {
		slog.Error("Data too big", "remote_addr", c.conn.RemoteAddr(), "size", size, "offset", offset, "block_offset", offset/common.BlockSize)
		c.respondError(fsapi.Error_ERROR_EMSGSIZE)
		return
	}

	file, ok := c.server.fsTable.getFile(write.GetId())
	if !ok {
		slog.Error("Cannot write to non-existent file", "remote_addr", c.conn.RemoteAddr())
		c.respondError(fsapi.Error_ERROR_ENOENT)
		return
	}

	_, ok = file.getBlock(write.GetOffset())

	if !ok {
		blk, err := c.server.storageList.createBlock(write.GetData())
		if err != 0 {
			c.respondError(err)
			return
		}
		file.addBlock(write.GetOffset(), blk)
		c.respondWrite()
		return
	}

	slog.Error("Not implemented: write to existing block", "remote_addr", c.conn.RemoteAddr())
	c.respondError(fsapi.Error_ERROR_ENOENT)
}

func (c *clientConn) handleRead(req *fsapi.Request) {
	read := req.GetRead()

	file, ok := c.server.fsTable.getFile(read.GetId())
	if !ok {
		slog.Error("Cannot read from non-existent file", "remote_addr", c.conn.RemoteAddr())
		c.respondError(fsapi.Error_ERROR_ENOENT)
		return
	}

	blk, ok := file.getBlock(read.GetOffset())
	if !ok {
		c.respondRead(make([]byte, 0))
		return
	}

	storageServer := c.server.storageList.getStorage(blk.storageIdx)
	size := min(read.GetSize(), blk.size)
	resp := storageServer.read(blk.id, size, read.GetOffset()%common.BlockSize)
	c.respondRead(resp.GetData())
}
