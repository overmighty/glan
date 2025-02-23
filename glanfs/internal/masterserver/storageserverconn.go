package masterserver

import (
	"github.com/overmighty/glan/glanfs/api/storageapi"
	"github.com/overmighty/glan/glanfs/internal/common"
	"log/slog"
	"sync"
)

type storageServerConn struct {
	server *MasterServer

	connMu sync.Mutex
	conn   *common.Conn[*storageapi.Request, *storageapi.Response]

	numFreeBlocks uint64
}

func (s *storageServerConn) do(req *storageapi.Request) *storageapi.Response {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if err := s.conn.WriteMessage(req); err != nil {
		slog.Error("Failed to write message", "err", err, "remote_addr", s.conn.RemoteAddr())
		panic(err)
	}

	var resp storageapi.Response
	//startTime := time.Now()
	err := s.conn.ReadMessage(&resp)
	//endTime := time.Now()
	//slog.Debug("storageServerConn", "wait_time_ns", endTime.Sub(startTime).Nanoseconds())
	if err != nil {
		slog.Error("Failed to read message", "err", err, "remote_addr", s.conn.RemoteAddr())
		panic(err)
	}
	return &resp
}

func (s *storageServerConn) getCapacity() *storageapi.GetCapacityResponse {
	var req storageapi.Request

	var body storageapi.GetCapacityRequest

	req.SetGetCapacity(&body)

	resp := s.do(&req)
	return resp.GetGetCapacity()
}

func (s *storageServerConn) write(blockID uint64, data []byte, offset uint64) *storageapi.WriteResponse {
	var req storageapi.Request

	var body storageapi.WriteRequest
	body.SetBlockId(blockID)
	body.SetData(data)
	body.SetOffset(offset)

	req.SetWrite(&body)

	resp := s.do(&req)
	return resp.GetWrite()
}

func (s *storageServerConn) read(blockID uint64, size uint64, offset uint64) *storageapi.ReadResponse {
	var req storageapi.Request

	var body storageapi.ReadRequest
	body.SetBlockId(blockID)
	body.SetSize(size)
	body.SetOffset(offset)

	req.SetRead(&body)

	resp := s.do(&req)
	return resp.GetRead()
}
