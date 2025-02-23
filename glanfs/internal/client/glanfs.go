package client

import (
	"context"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/overmighty/glan/glanfs/internal/common"
	"log/slog"
	"syscall"
)

type glanfsNode struct {
	fs.Inode

	client *Client
	id     uint64

	size uint64
}

func (n *glanfsNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	slog.Debug("Lookup called", "name", name)

	resp, err := n.client.lookupNode(n.id, name)
	if err != 0 {
		return nil, err
	}

	child := &glanfsNode{
		client: n.client,
		id:     resp.GetId(),
	}
	id := fs.StableAttr{
		Mode: fsapiFileTypeToMode(resp.GetType()),
		Ino:  resp.GetId(),
	}
	out.Mode = 0700
	return n.NewInode(ctx, child, id), 0
}

func (n *glanfsNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	slog.Debug("Create called", "name", name, "flags", flags, "mode", mode)

	resp, err := n.client.createNode(n.id, name)
	if err != 0 {
		return nil, nil, 0, err
	}

	child := &glanfsNode{
		client: n.client,
		id:     resp.GetId(),
	}
	id := fs.StableAttr{
		Mode: mode,
		Ino:  resp.GetId(),
	}
	fh = &glanfsFileHandle{
		client: n.client,
		id:     resp.GetId(),
	}
	out.Mode = 0777
	return n.NewInode(ctx, child, id), fh, 0, 0
}

func (n *glanfsNode) Readdir(context.Context) (fs.DirStream, syscall.Errno) {
	slog.Debug("Readdir called")

	resp, err := n.client.readdirNode(n.id)
	if err != 0 {
		return nil, err
	}

	var dirEntries []fuse.DirEntry
	for _, entry := range resp.GetEntries() {
		d := fuse.DirEntry{
			Mode: fsapiFileTypeToMode(entry.GetType()),
			Name: entry.GetName(),
			Ino:  entry.GetId(),
		}
		dirEntries = append(dirEntries, d)
	}
	return fs.NewListDirStream(dirEntries), 0
}

func (n *glanfsNode) Open(_ context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	slog.Debug("Open called", "flags", flags)

	return &glanfsFileHandle{n.client, n.id}, 0, 0
}

func (n *glanfsNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	slog.Debug("Getattr called")

	resp, err := n.client.getattrNode(n.id)
	if err != 0 {
		return err
	}

	out.AttrValid = 1

	out.Ino = n.id
	out.Size = resp.GetSize()

	out.Blksize = common.BlockSize

	return 0
}

var _ = (fs.InodeEmbedder)((*glanfsNode)(nil))
var _ = (fs.NodeLookuper)((*glanfsNode)(nil))
var _ = (fs.NodeCreater)((*glanfsNode)(nil))
var _ = (fs.NodeReaddirer)((*glanfsNode)(nil))
var _ = (fs.NodeOpener)((*glanfsNode)(nil))
var _ = (fs.NodeGetattrer)((*glanfsNode)(nil))

type glanfsFileHandle struct {
	client *Client
	id     uint64
}

func (f *glanfsFileHandle) Write(_ context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	slog.Debug("Write called", "size", len(data), "off", off)

	if off < 0 {
		panic(fmt.Errorf("negative off: %d", off))
	}

	totalSize := uint64(len(data))
	var totalWritten uint32

	var i uint64
	for i < totalSize {
		offset := uint64(off) + i
		offsetInBlock := offset % common.BlockSize

		size := min(totalSize-i, common.BlockSize-offsetInBlock)

		_, err := f.client.writeFile(f.id, data[i:i+size], offset)
		if err != 0 {
			return totalWritten, err
		}
		totalWritten += uint32(size)

		i += size
	}

	return totalWritten, 0
}

type readResult struct {
	data []byte
}

func (r *readResult) Bytes([]byte) ([]byte, fuse.Status) {
	return r.data, fuse.OK
}

func (r *readResult) Size() int {
	return len(r.data)
}

func (r *readResult) Done() {
	//slog.Debug("Read done")
}

func (f *glanfsFileHandle) Read(_ context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	slog.Debug("Read called", "size", len(dest), "off", off)

	if off < 0 {
		panic(fmt.Errorf("negative off: %d", off))
	}

	totalSize := uint64(len(dest))

	var i uint64
	for i < totalSize {
		offset := uint64(off) + i
		offsetInBlock := offset % common.BlockSize

		size := min(totalSize-i, common.BlockSize-offsetInBlock)

		resp, err := f.client.readFile(f.id, size, offset)
		if err != 0 {
			return nil, err
		}
		copy(dest[i:i+size], resp.GetData())

		i += size
	}

	return &readResult{data: dest}, 0
}

var _ = (fs.FileWriter)((*glanfsFileHandle)(nil))
var _ = (fs.FileReader)((*glanfsFileHandle)(nil))
