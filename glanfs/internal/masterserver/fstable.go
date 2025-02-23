package masterserver

import (
	"github.com/overmighty/glan/glanfs/api/fsapi"
	"sync"
)

type fsTable struct {
	mu         sync.RWMutex
	nextNodeID uint64
	nodesByID  map[uint64]fsNode
}

func newFSTable() *fsTable {
	return &fsTable{
		nextNodeID: 1,
		nodesByID:  make(map[uint64]fsNode),
	}
}

func (t *fsTable) addNode(parentID uint64, node fsNode) fsapi.Error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if parentID != 0 {
		parentNode, ok := t.nodesByID[parentID]
		if !ok {
			return fsapi.Error_ERROR_ENOENT
		}

		parentDir, ok := parentNode.(*dirNode)
		if !ok {
			return fsapi.Error_ERROR_ENOTDIR
		}

		name := node.getName()
		if _, ok = parentDir.getEntry(name); ok {
			return fsapi.Error_ERROR_EEXIST
		}
		parentDir.addEntry(node)
	}

	node.setID(t.nextNodeID)
	t.nextNodeID++

	t.nodesByID[node.getID()] = node
	return 0
}

func (t *fsTable) getNode(id uint64) (node fsNode, ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node, ok = t.nodesByID[id]
	return
}

func (t *fsTable) getDir(id uint64) (dir *dirNode, ok bool) {
	node, ok := t.getNode(id)
	if !ok {
		return nil, false
	}

	dir, ok = node.(*dirNode)
	if !ok {
		return nil, false
	}

	return dir, true
}

func (t *fsTable) getFile(id uint64) (file *fileNode, ok bool) {
	node, ok := t.getNode(id)
	if !ok {
		return nil, false
	}

	file, ok = node.(*fileNode)
	if !ok {
		return nil, false
	}

	return file, true
}
