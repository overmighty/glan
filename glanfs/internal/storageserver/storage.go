package storageserver

import (
	"github.com/overmighty/glan/glanfs/internal/common"
	"sync"
)

type storage struct {
	mu         sync.RWMutex
	blocksByID map[uint64][]byte
}

func newStorage() *storage {
	return &storage{
		blocksByID: make(map[uint64][]byte),
	}
}

func (t *storage) write(blockID uint64, data []byte, offset uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if offset+uint64(len(data)) > common.BlockSize {
		panic("data too big")
	}

	blk, ok := t.blocksByID[blockID]
	if !ok {
		blk = make([]byte, common.BlockSize)
		t.blocksByID[blockID] = blk
	}
	copy(blk[offset:], data)
}

func (t *storage) read(blockID uint64, size uint64, offset uint64) []byte {
	t.mu.RLock()
	defer t.mu.RUnlock()

	blk, ok := t.blocksByID[blockID]
	if !ok {
		panic("block not found")
	}
	data := make([]byte, size)
	copy(data, blk[offset:offset+size])
	return data
}
