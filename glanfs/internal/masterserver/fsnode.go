package masterserver

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"sync"
)

type fsNode interface {
	getType() fsapi.FileType
	getID() uint64
	setID(id uint64)
	getName() string
}

type dirNode struct {
	mu            sync.RWMutex
	id            uint64
	name          string
	entriesByName map[string]fsNode
}

func newDirNode(name string) *dirNode {
	return &dirNode{
		name:          name,
		entriesByName: make(map[string]fsNode),
	}
}

func (d *dirNode) getType() fsapi.FileType {
	return fsapi.FileType_FILE_TYPE_DIRECTORY
}

func (d *dirNode) getID() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.id
}

func (d *dirNode) setID(id uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.id = id
}

func (d *dirNode) getName() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.name
}

func (d *dirNode) getEntry(name string) (fsNode, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entry, ok := d.entriesByName[name]
	return entry, ok
}

func (d *dirNode) addEntry(entry fsNode) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entriesByName[entry.getName()] = entry
}

type fileNode struct {
	mu             sync.RWMutex
	id             uint64
	name           string
	blocksByOffset map[uint64]*block
}

func newFileNode(name string) *fileNode {
	return &fileNode{
		name:           name,
		blocksByOffset: make(map[uint64]*block),
	}
}

func (f *fileNode) getType() fsapi.FileType {
	return fsapi.FileType_FILE_TYPE_REGULAR
}

func (f *fileNode) getID() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.id
}

func (f *fileNode) setID(id uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.id = id
}

func (f *fileNode) getName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.name
}

func (f *fileNode) getBlock(dataOffset uint64) (*block, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	blockOffset := dataOffset / common.BlockSize
	blk, ok := f.blocksByOffset[blockOffset]
	return blk, ok
}

func (f *fileNode) addBlock(dataOffset uint64, blk *block) {
	f.mu.Lock()
	defer f.mu.Unlock()

	blockOffset := dataOffset / common.BlockSize
	f.blocksByOffset[blockOffset] = blk
}

func (f *fileNode) size() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var size uint64
	for _, blk := range f.blocksByOffset {
		size += blk.size
	}
	return size
}
