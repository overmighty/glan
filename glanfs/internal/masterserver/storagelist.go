package masterserver

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"log/slog"
	"math/rand/v2"
	"sync"
)

type storageList struct {
	mu              sync.RWMutex
	rnd             *rand.Rand
	nextBlockID     uint64
	storageServers  []*storageServerConn
	totalFreeBlocks uint64
}

func newStorageList() *storageList {
	return &storageList{
		rnd: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

func (l *storageList) addStorage(s *storageServerConn) {
	l.mu.Lock()
	defer l.mu.Unlock()

	slog.Info("Adding storage server", "remote_addr", s.conn.RemoteAddr(), "capacity_blocks", s.numFreeBlocks)
	l.storageServers = append(l.storageServers, s)

	l.totalFreeBlocks += s.numFreeBlocks
}

func (l *storageList) getStorage(idx int) *storageServerConn {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storageServers[idx]
}

func (l *storageList) createBlock(data []byte) (*block, fsapi.Error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.totalFreeBlocks == 0 {
		return nil, fsapi.Error_ERROR_ENOSPC
	}

	for {
		storageIdx := l.rnd.IntN(len(l.storageServers))
		if l.storageServers[storageIdx].numFreeBlocks == 0 {
			continue
		}

		s := l.storageServers[storageIdx]

		id := l.nextBlockID
		l.nextBlockID++

		blk := &block{
			id:         id,
			size:       uint64(len(data)),
			storageIdx: storageIdx,
		}
		_ = s.write(blk.id, data, 0)

		s.numFreeBlocks -= 1
		l.totalFreeBlocks -= 1

		return blk, 0
	}
}
